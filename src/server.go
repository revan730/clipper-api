package src

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/rs/cors"

	"github.com/revan730/clipper-api/db"
	"github.com/revan730/clipper-api/types"
	"github.com/revan730/clipper-api/queue"
	"github.com/revan730/clipper-api/CIApi"
	commonTypes "github.com/revan730/clipper-common/types"
)

// Server holds application API server
type Server struct {
	logger         *zap.Logger
	config         *types.Config
	databaseClient db.DatabaseClient
	ciClient *CIApi.CIClient
	jobQueue *queue.Queue
	router         *gin.Engine
}

// NewServer creates new copy of Server
func NewServer(logger *zap.Logger, config *types.Config) *Server {
	server := &Server{
		logger: logger,
		router: gin.Default(),
		config: config,
	}
	server.jobQueue = queue.NewQueue(config.RabbitAddress)
	dbConfig := types.PGClientConfig{
		DBUser:         config.DBUser,
		DBAddr:         config.DBAddr,
		DBPassword:     config.DBPassword,
		DB:     config.DB,
		AdminLogin: config.AdminLogin,
		AdminPassword: config.AdminPassword,
	}
	dbClient := db.NewPGClient(dbConfig)
	server.databaseClient = dbClient
	ciClient := CIApi.NewClient(config.CIAddress, logger)
	server.ciClient = ciClient
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
		authorized.POST("/api/v1/user/accessToken", s.setAccessTokenHandler)
		// Github repos
		authorized.POST("/api/v1/repos", s.postRepoHandler)
		authorized.GET("/api/v1/repos", s.getAllReposHandler)
		authorized.GET("/api/v1/repos/:id", s.getRepoHandler)
		authorized.DELETE("/api/v1/repos/:id", s.deleteRepoHandler)
		// Repo configs
		authorized.POST("/api/v1/repos/:id/branch", s.postBranchConfigHandler)
		authorized.GET("/api/v1/repos/:id/branch", s.getAllBranchConfigsHandler)
		authorized.DELETE("/api/v1/repos/:id/branch/:branch", s.deleteBranchConfigHandler)
		// Builds
		authorized.GET("/api/v1/builds/:id", s.getBuildHandler)
		// Build artifacts
		authorized.GET("/api/v1/builds/:id/artifact", s.getBuildArtifactHandler)
		// TODO: /api/v1/repos/:id/builds GET
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

func (s *Server) startCIJob(ciMsg commonTypes.CIJob) error {
	s.logInfo("Starting CI Job")
	return s.jobQueue.PublishCIJob(&ciMsg)
}