package src

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
	"strconv"
	"strings"
	"errors"
	"encoding/hex"
	"crypto/hmac"
	"crypto/sha1"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	jwt "github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/rs/cors"

	"github.com/go-redis/redis"
	"github.com/revan730/diploma-server/db"
	"github.com/revan730/diploma-server/types"
)

// Server holds application API server
type Server struct {
	logger         *zap.Logger
	config         *types.Config
	redisClient    *redis.Client
	databaseClient *db.DatabaseClient
	router         *gin.Engine
}

func jwtMiddleware(secret []byte) gin.HandlerFunc {
	// TODO: Json error handler
	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	return func(c *gin.Context) {
		err := jwtMiddleware.CheckJWT(c.Writer, c.Request)
		if err != nil {
			return
		}
		c.Next()
	}
}

func (s *Server) getClaimByName(c *gin.Context, name string) interface{} {
	jwtToken := c.Request.Context().Value("user")
	claims := jwtToken.(*jwt.Token).Claims.(jwt.MapClaims)
	return claims[name]
}

func (s *Server) getUserLoginClaim(c *gin.Context) (string, bool) {
	login, ok := s.getClaimByName(c, "user").(string)
	if ok == false {
		s.logInfo("Failed to get user login claim")
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return "", false
	}
	return login, true
}

func (s *Server) getUserIDClaim(c *gin.Context) (int64, bool) {
	userID, ok := s.getClaimByName(c, "userID").(float64)
	if ok == false {
		s.logInfo("Failed to get userID claim")
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return 0, false
	}
	return int64(userID), true
}

func (s *Server) getUserAdminClaim(c *gin.Context) (bool, bool) {
	isAdmin, ok := s.getClaimByName(c, "admin").(bool)
	if ok == false {
		s.logInfo("Failed to get user admin claim")
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return false, false
	}
	return isAdmin, true
}

func (s *Server) userClaimMiddleware(c *gin.Context) {
	userID, ok := s.getUserIDClaim(c)
	if ok == false {
		return
	}
	login, ok := s.getUserLoginClaim(c)
	if ok == false {
		return
	}
	isAdmin, ok := s.getUserAdminClaim(c)
	if ok == false {
		return
	}
	user := types.User{
		ID: userID,
		IsAdmin: isAdmin,
		Login: login,
	}
	c.Set("userClaim", user)
	c.Next()
}

// NewServer creates new copy of Server
func NewServer(logger *zap.Logger, config *types.Config) *Server {
	server := &Server{
		logger: logger,
		router: gin.Default(),
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
	dbClient := db.NewDBClient(*config)
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

// Routes binds api routes to handlers
func (s *Server) Routes() *Server {
	s.router.POST("/api/v1/login", s.loginHandler)
	s.router.POST("/api/v1/register", s.registerHandler)

	s.router.POST("/api/v1/octohook/:user", s.webhookHandler)

	jwtSecret := []byte(s.config.JWTSecret)
	authorized := s.router.Group("/")
	authorized.Use(jwtMiddleware(jwtSecret), s.userClaimMiddleware)
	{
		// User
		authorized.POST("/api/v1/user/secret", s.setSecretHandler)
		// Github repos
		authorized.POST("/api/v1/repos", s.postRepoHandler)
		authorized.GET("/api/v1/repos", s.getAllReposHandler)
		authorized.GET("/api/v1/repos/:id", s.getRepoHandler)
		authorized.DELETE("/api/v1/repos/:id", s.deleteRepoHandler)
		// Repo configs
		authorized.POST("/api/v1/repos/:id/branch", s.postBranchConfigHandler)
		authorized.GET("/api/v1/repos/:id/branch", s.getAllBranchConfigsHandler)
		authorized.DELETE("/api/v1/repos/:id/branch/:branch", s.deleteBranchConfigHandler)
	}
	return s
}

// Run starts api server
func (s *Server) Run() {
	defer s.databaseClient.Close()
	rand.Seed(time.Now().UnixNano())
	err := s.databaseClient.CreateSchema()
	if err != nil {
		s.logError("Failed to create database schema", err)
		os.Exit(1)
	}
	s.logger.Info("Starting server", zap.Int("port", s.config.Port))
	corsRouter := cors.Default().Handler(s.router)
	err = http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), corsRouter)
	if err != nil {
		s.logError("Server failed", err)
		os.Exit(1)
	}
}

func (s *Server) bindJSON(c *gin.Context, msg interface{}) bool {
	err := c.ShouldBindJSON(&msg)
	if err != nil {
		s.logError("JSON read error", err)
		c.JSON(http.StatusBadRequest, gin.H{"err": "Bad json"})
		return false
	}
	return true
}

func (s *Server) loginHandler(c *gin.Context) {
	// Check if login and password are provided
	loginMsg := &types.CredentialsMessage{}
	s.bindJSON(c, loginMsg)
	// TODO: Check bindJSON result
	if loginMsg.Login == "" || loginMsg.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Empty login or password"})
		return
	}
	user, err := s.databaseClient.FindUser(loginMsg.Login)
	if err != nil {
		s.logError("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Failed to login"})
		return
	}
	if user.Authenticate(loginMsg.Password) == false {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "Failed to login"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": user.IsAdmin,
		"user": user.Login,
		"userID": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		s.logError("jwt error", err)
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (s *Server) registerHandler(c *gin.Context) {
	// Check if login and password are provided
	registerMsg := &types.CredentialsMessage{}
	s.bindJSON(c, registerMsg)
	if registerMsg.Login == "" || registerMsg.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Empty login or password"})
		return
	}
	err := s.databaseClient.CreateUser(registerMsg.Login, registerMsg.Password, false)
	if err != nil {
		// TODO: Maybe move this error handling to CreateUser func?
		pgErr, ok := err.(pg.Error)
		if ok && pgErr.IntegrityViolation() {
			c.JSON(http.StatusBadRequest, gin.H{"err": "User already exists"})
			return
		}
		s.logError("Create user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) setSecretHandler(c *gin.Context) {
	secretMsg := &types.WebhookSecretMessage{}
	bound := s.bindJSON(c, secretMsg)
	if bound == false {
		return
	}
	if secretMsg.Secret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "secret not provided"})
		return
	}
	userClaim := c.MustGet("userClaim").(types.User)
	user, err := s.databaseClient.FindUser(userClaim.Login)
	if err != nil {
		s.logError("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "user not found"})
		return
	}
	user.WebhookSecret = secretMsg.Secret
	err = s.databaseClient.SaveUser(user)
	if err != nil {
		s.logError("User save error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) postRepoHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoMsg := &types.RepoMessage{}
	bound := s.bindJSON(c, repoMsg)
	if bound == false {
		return
	}
	if repoMsg.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo name not provided"})
		return
	}
	err := s.databaseClient.CreateRepo(repoMsg.FullName, userClaim.ID)
	if err != nil {
		// TODO: Maybe move this error handling to CreateUser func?
		pgErr, ok := err.(pg.Error)
		if ok && pgErr.IntegrityViolation() {
			c.JSON(http.StatusBadRequest, gin.H{"err": "Repo already exists"})
			return
		}
		s.logError("Create repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) getRepoHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not provided"})
		return
	}
	if repo.UserID != userClaim.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "you have no access to this repo"})
	}
	c.JSON(http.StatusOK, repo)
}

func (s *Server) getAllReposHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repos, err := s.databaseClient.FindAllUserRepos(userClaim.ID, c.Request.URL.Query())
	if err != nil {
		s.logError("Find repos error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"repos": repos})
}

func (s *Server) deleteRepoHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	// TODO: Don't think this part is effective, as user who
	// forges access tokens can simply know user's id from
	// token itself, no need to guess
	user, err := s.databaseClient.FindUserByID(userClaim.ID)
	if err != nil {
		s.logError("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		return
	}
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if user.IsAdmin == false {
		if userClaim.ID != repo.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		    return
		}
	}
	// TODO: Explicitly handle missing repo error
	err = s.databaseClient.DeleteRepoByID(int64(repoID))
	if err != nil {
		s.logError("Delete repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) postBranchConfigHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	user, err := s.databaseClient.FindUserByID(userClaim.ID)
	if err != nil {
		s.logError("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		return
	}
	// TODO: Probably this part must be overwritten as pg
	// will return error on attemt to create config with non-existing
	// repo id
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if user.IsAdmin == false {
		if userClaim.ID != repo.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		    return
		}
	}
	branchMsg := &types.BranchMessage{}
	bound := s.bindJSON(c, branchMsg)
	if bound == false {
		return
	}
	if branchMsg.BranchName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "branch name not provided"})
		return
	}
	// Check if config for this branch already exists for this repo
	conf, err := s.databaseClient.FindBranchConfig(int64(repoID), branchMsg.BranchName)
	if err != nil {
		s.logError("Find branch config error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conf != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Config for this branch already exists"})
		return
	}
	branchConf := types.BranchConfig{
		GithubRepoID: int64(repoID),
		Branch: branchMsg.BranchName,
		IsCiEnabled: true,
	}
	err = s.databaseClient.CreateBranchConfig(&branchConf)
	if err != nil {
		s.logError("Create branch config error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err":nil})
}

func (s *Server) getAllBranchConfigsHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	// Check if user owns this repo
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if userClaim.IsAdmin == false {
		if userClaim.ID != repo.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		    return
		}
	}
	configs, err := s.databaseClient.FindAllBranchConfigs(repo.ID, c.Request.URL.Query())
	if err != nil {
		s.logError("Find branch configs error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func (s *Server) deleteBranchConfigHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	branchName := c.Param("branch")
	// Check if user owns this repo
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if userClaim.IsAdmin == false {
		if userClaim.ID != repo.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		    return
		}
	}
	err = s.databaseClient.DeleteBranchConfig(int64(repoID), branchName)
	// TODO: Explicitly handle missing branch config error
	if err != nil {
		s.logError("Delete branch config error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) startCIJob() {
	s.logInfo("Starting CI Job")
}

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}


func checkSecret(secret string, c *gin.Context) error {
	if secret == "" {
		return nil
	}
	gitSignStr := c.GetHeader("X-Hub-Signature")
	if gitSignStr == "" {
		return errors.New("Github signature not provided")
	}
	fmt.Println(gitSignStr)
	rawMsg, ok := c.Get(gin.BodyBytesKey)
	if ok != true {
		return errors.New("Failed to get request body")
	}
	body, ok := rawMsg.([]byte)
	if ok != true {
		return errors.New("Failed to assert request body")
	}
	actualSign := make([]byte, 20)
	hex.Decode(actualSign, []byte(gitSignStr[5:]))
	fmt.Println(actualSign)

	if hmac.Equal(signBody([]byte(secret), body), actualSign) == false {
		return errors.New("Signature doesn't match")
	}
	return nil
}

func (s *Server) webhookHandler(c *gin.Context) {
	payload := &types.WebhookMessage{}
	err := c.ShouldBindBodyWith(&payload, binding.JSON)
	if err != nil {
		s.logError("JSON read error", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(payload)
	// Get user by user param
	userLogin := c.Param("user")
	user, err := s.databaseClient.FindUser(userLogin)
	if err != nil {
		s.logError("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = checkSecret(user.WebhookSecret, c)
	if err != nil {
		s.logError("Webhook secret error", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	eventType := c.GetHeader("X-GitHub-Event")
	switch eventType {
	case "push":
		repo, err := s.databaseClient.FindRepoByName(payload.Repository.FullName)
		if err != nil {
			s.logError("Failed to find repo", err)
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		branchName := strings.Split(payload.Ref, "/")[2]
		config, err := s.databaseClient.FindBranchConfig(repo.ID, branchName)
		if err != nil {
			s.logError("Failed to find branch config", err)
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Automatic CI is not enabled for this branch, ignore
		if config == nil || config.IsCiEnabled == false {
			c.Writer.WriteHeader(http.StatusOK)
			return
		}
		// TODO: Start CI Job
		s.startCIJob()
		c.Writer.WriteHeader(http.StatusOK)
		return
	default:
		s.logInfo("Unsupported type, ignoring")
		c.Writer.WriteHeader(http.StatusOK)
		return
	// TODO: Implement pull_request event handling
	}
}