package src

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-pg/pg"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/go-redis/redis"
	"github.com/revan730/diploma-server/db"
	"github.com/revan730/diploma-server/types"
)

type Server struct {
	logger         *zap.Logger
	config         *Config
	redisClient    *redis.Client
	databaseClient *db.DatabaseClient
	router         *httprouter.Router
}

func NewServer(logger *zap.Logger, config *Config) *Server {
	server := &Server{
		logger: logger,
		router: httprouter.New(),
		config: config,
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	dbClient := db.NewDBClient(config.DBAddr, config.DB, config.DBUser, config.DBPassword)
	server.redisClient = redisClient
	server.databaseClient = dbClient
	return server
}

func (s *Server) logError(msg string, err error) {
	defer s.logger.Sync()
	s.logger.Error(msg, zap.String("packageLevel", "core"), zap.Error(err))
}

func (s *Server) logInfo(msg string) {
	defer s.logger.Sync()
	s.logger.Info("INFO", zap.String("msg", msg), zap.String("packageLevel", "core"))
}

func (s *Server) Routes() *Server {
	s.router.POST("/api/v1/login", s.LoginHandler)
	s.router.POST("/api/v1/register", s.RegisterHandler)
	return s
}

func writeJSON(w http.ResponseWriter, d interface{}) {
	j, _ := json.Marshal(d)
	fmt.Fprint(w, string(j))
}

func readJSON(body io.ReadCloser, jtype interface{}) error {
	// Read body
	if body == nil {
		return errors.New("Body is nil")
	}
	b, err := ioutil.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}

	// Decode json into provided structure
	return json.Unmarshal(b, jtype)

}

func (s *Server) writeResponse(w http.ResponseWriter, responseBody interface{}, responseCode int) {
	w.WriteHeader(responseCode)
	writeJSON(w, responseBody)
}

func (s *Server) Run() {
	defer s.databaseClient.Close()
	rand.Seed(time.Now().UnixNano())
	err := s.databaseClient.CreateSchema()
	if err != nil {
		s.logError("Failed to create database schema", err)
		os.Exit(1)
	}
	s.logger.Info("Starting server", zap.Int("port", s.config.Port))
	err = http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), s.router)
	if err != nil {
		s.logError("Server failed", err)
		os.Exit(1)
	}
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Check if login and password are provided
	var loginMsg types.CredentialsMessage
	err := readJSON(r.Body, &loginMsg)
	if err != nil {
		s.logError("JSON read error", err)
		s.writeResponse(w, &map[string]string{"err": "Bad json"}, http.StatusBadRequest)
		return
	}
	if loginMsg.Login == "" || loginMsg.Password == "" {
		s.writeResponse(w, &map[string]string{"err": "Empty login or password"}, http.StatusBadRequest)
		return
	}
	user, err := s.databaseClient.FindUser(loginMsg.Login)
	if err != nil {
		s.logError("Find user error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		s.writeResponse(w, &map[string]string{"err": "Failed to login"}, http.StatusUnauthorized)
		return
	}
	if user.Authenticate(loginMsg.Password) == false {
		s.writeResponse(w, &map[string]string{"err": "Failed to login"}, http.StatusUnauthorized)
		return
	}

	// Generate user token and save to redis
	tokenBytes := make([]byte, 8)
	rand.Read(tokenBytes)
	authToken := base64.StdEncoding.EncodeToString(tokenBytes)
	s.redisClient.Set(authToken, user.Id, 6*time.Hour).Result()

	s.writeResponse(w, &map[string]string{"token": authToken}, http.StatusOK)
}

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Check if login and password are provided
	var registerMsg types.CredentialsMessage
	err := readJSON(r.Body, &registerMsg)
	if err != nil {
		s.logError("JSON read error", err)
		s.writeResponse(w, &map[string]string{"err": "Bad json"}, http.StatusBadRequest)
		return
	}
	if registerMsg.Login == "" || registerMsg.Password == "" {
		s.writeResponse(w, &map[string]string{"err": "Empty login or password"}, http.StatusBadRequest)
		return
	}
	err = s.databaseClient.CreateUser(registerMsg.Login, registerMsg.Password)
	if err != nil {
		// TODO: Maybe move this error handling to CreateUser func?
		pgErr, ok := err.(pg.Error)
		if ok && pgErr.IntegrityViolation() {
			s.writeResponse(w, &map[string]string{"err": "User already exists"}, http.StatusBadRequest)
			return
		}
		s.logError("Create user error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
