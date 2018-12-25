package src

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"

	"github.com/revan730/clipper-api/db"
	"github.com/revan730/clipper-api/log"
	"github.com/revan730/clipper-api/types"
	"github.com/revan730/clipper-api/queue"
	"github.com/revan730/clipper-api/CIApi"
	"github.com/revan730/clipper-api/CDApi"
	commonTypes "github.com/revan730/clipper-common/types"
)

// Server holds application API server
type Server struct {
	log         log.Logger
	config         *types.Config
	databaseClient db.DatabaseClient
	ciClient *CIApi.CIClient
	cdClient *CDApi.CDClient
	jobQueue queue.Queue
	router         *gin.Engine
}

// NewServer creates new copy of Server
func NewServer(logger log.Logger, config *types.Config) *Server {
	server := &Server{
		log: logger,
		router: gin.Default(),
		config: config,
	}
	server.jobQueue = queue.NewRMQQueue(config.RabbitAddress)
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
	cdClient := CDApi.NewClient(config.CDAddress, logger)
	server.cdClient = cdClient
	return server
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
		authorized.GET("/api/v1/repos/:id/builds", s.getAllBuildsHandler)
		// Builds
		authorized.GET("/api/v1/builds/:id", s.getBuildHandler)
		// Build artifacts
		authorized.GET("/api/v1/builds/:id/artifact", s.getBuildArtifactHandler)
		// Deployments
		authorized.POST("/api/v1/deployments", s.postDeploymentHandler)
	}
	return s
}

// Run starts api server
func (s *Server) Run() {
	defer s.databaseClient.Close()
	rand.Seed(time.Now().UnixNano())
	err := s.databaseClient.CreateSchema()
	if err != nil {
		s.log.Error("Failed to create database schema", err)
		os.Exit(1)
	}
	s.log.Info(fmt.Sprintf("Starting server at port %d", s.config.Port))
	corsRouter := cors.Default().Handler(s.router)
	err = http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), corsRouter)
	if err != nil {
		s.log.Error("Server failed", err)
		os.Exit(1)
	}
}

func (s *Server) bindJSON(c *gin.Context, msg interface{}) bool {
	err := c.ShouldBindJSON(&msg)
	if err != nil {
		s.log.Error("JSON read error", err)
		c.JSON(http.StatusBadRequest, gin.H{"err": "Bad json"})
		return false
	}
	return true
}

func (s *Server) startCIJob(ciMsg commonTypes.CIJob) error {
	s.log.Info("Starting CI Job")
	return s.jobQueue.PublishCIJob(&ciMsg)
}