package src

import (
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
	jwt "github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"

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

func JWTMiddlewareFromHandler(handlerF http.HandlerFunc, secret []byte) http.HandlerFunc {
	// TODO: Json error handler
	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := jwtMiddleware.CheckJWT(w, r)

		if err != nil {
			return
		}
		handlerF(w, r)
	})
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
	dbClient := db.NewDBClient(config.DBAddr, config.DB, config.DBUser, config.DBPassword, config.AdminLogin, config.AdminPassword)
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
	jwtSecret := []byte(s.config.JWTSecret)
	s.router.POST("/api/v1/login", s.LoginHandler)
	s.router.POST("/api/v1/register", s.RegisterHandler)
	s.router.HandlerFunc("POST","/api/v1/setsecret", JWTMiddlewareFromHandler(s.SetSecretHandler, jwtSecret))
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": user.IsAdmin,
		"user": user.Login,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		s.logError("jwt error", err)
	}

	s.writeResponse(w, &map[string]string{"token": tokenString}, http.StatusOK)
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
	err = s.databaseClient.CreateUser(registerMsg.Login, registerMsg.Password, false)
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
	s.writeResponse(w, &map[string]interface{}{"err": nil}, http.StatusOK)
}

func (s *Server) SetSecretHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	// Hint: token claims can be retreived from context
	var secretMsg types.WebhookSecretMessage
	err := readJSON(r.Body, &secretMsg)
	if err != nil {
		s.logError("JSON read error", err)
		s.writeResponse(w, &map[string]string{"err": "Bad json"}, http.StatusBadRequest)
		return
	}
	if secretMsg.Secret == "" {
		s.writeResponse(w, &map[string]string{"err": "secret not provided"}, http.StatusBadRequest)
		return
	}
	// TODO: Extract function to get claim by name
	jwtToken := r.Context().Value("user")
	claims := jwtToken.(*jwt.Token).Claims.(jwt.MapClaims)
	login := claims["user"].(string)
	user, err := s.databaseClient.FindUser(login)
	if err != nil {
		s.logError("Find user error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		s.writeResponse(w, &map[string]string{"err": "user not found"}, http.StatusUnauthorized)
		return
	}
	user.WebhookSecret = secretMsg.Secret
	err = s.databaseClient.SaveUser(user)
	if err != nil {
		s.logError("User save error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.writeResponse(w, &map[string]interface{}{"err": nil}, http.StatusOK)
}
