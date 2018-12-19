package src

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-api/types"
)

func (s *Server) postDeploymentHandler(c *gin.Context) {
	deploymentMsg := &types.DeploymentMessage{}
	bound := s.bindJSON(c, deploymentMsg)
	if bound != true {
		c.JSON(http.StatusBadRequest, gin.H{"err": "bad json"})
		return
	}
	// TODO: Check if repo and artifact belongs to user
	err := s.cdClient.CreateDeployment(deploymentMsg)
	if err != nil {
		s.logError("Create deployment error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}
