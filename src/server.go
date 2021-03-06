package src

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"

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
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Authorization",
	 "Content-Type",
	}
	s.router.Use(cors.New(corsConfig))
	s.router.POST("/api/v1/login", s.loginHandler)
	s.router.POST("/api/v1/register", s.registerHandler)

	s.router.POST("/api/v1/octohook/:user", s.webhookHandler)

	jwtSecret := []byte(s.config.JWTSecret)
	authorized := s.router.Group("/api/v1/")
	authorized.Use(s.jwtMiddleware(jwtSecret), s.userClaimMiddleware)
	{
		// User
		authorized.POST("user/secret", s.setSecretHandler)
		authorized.POST("user/accessToken", s.setAccessTokenHandler)
		// Github repos
		authorized.POST("repos", s.postRepoHandler)
		authorized.GET("repos", s.getAllReposHandler)
		authorized.GET("repos/:id", s.getRepoHandler)
		authorized.DELETE("repos/:id", s.deleteRepoHandler)
		// Repo configs
		authorized.POST("repos/:id/branch", s.postBranchConfigHandler)
		authorized.GET("repos/:id/branch", s.getAllBranchConfigsHandler)
		authorized.DELETE("repos/:id/branch/:branch", s.deleteBranchConfigHandler)
		// Builds
		authorized.GET("repos/:id/builds", s.getAllBuildsHandler)
		authorized.GET("builds/:id", s.getBuildHandler)
		// Build artifacts
		authorized.GET("builds/:id/artifact", s.getBuildArtifactHandler)
		authorized.GET("repos/:id/artifacts", s.getAllArtifactsHandler)

		admin := authorized.Group("admin")
		admin.Use(s.userIsAdminMiddleware)
		{
			// User control
			admin.GET("user", s.getAllUsersHandler)
			admin.POST("user/:id/admin", s.changeUserAdminHandler)
			// Deployments
			admin.GET("deployments", s.getAllDeploymentsHandler)
			admin.POST("deployments",s.postDeploymentHandler)
			admin.GET("deployments/:id", s.getDeploymentHandler)
			admin.DELETE("deployments/:id", s.deleteDeploymentHandler)
			admin.POST("deployments/:id/image", s.changeDeploymentImageHandler)
			admin.POST("deployments/:id/scale", s.scaleDeploymentHandler)
			admin.POST("deployments/:id/manifest", s.updateManifestHandler)
			// Revisions
			admin.GET("deployments/:id/revisions", s.getRevisionsHandler)
			admin.GET("revisions/:id", s.getRevisionHandler)
		}

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
	err = http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), s.router)
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