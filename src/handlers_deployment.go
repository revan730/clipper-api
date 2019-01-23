package src

import (
	"net/http"
	"strconv"

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
		s.log.Error("Create deployment error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) deleteDeploymentHandler(c *gin.Context) {
	depIDStr := c.Param("id")
	depID, err := strconv.ParseInt(depIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "deployment id is not int"})
		return
	}
	// TODO: Check if deployment belongs to user ?
	err = s.cdClient.DeleteDeployment(depID)
	if err != nil {
		s.log.Error("Delete deployment error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) changeDeploymentImageHandler(c *gin.Context) {
	depIDStr := c.Param("id")
	depID, err := strconv.ParseInt(depIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "deployment id is not int"})
		return
	}
	imageMsg := &types.ImageMessage{}
	bound := s.bindJSON(c, imageMsg)
	if bound != true {
		c.JSON(http.StatusBadRequest, gin.H{"err": "bad json"})
		return
	}
	rpcMsg := &types.DeploymentMessage{
		ID:         depID,
		ArtifactID: imageMsg.ImageID,
	}
	err = s.cdClient.UpdateImage(rpcMsg)
	if err != nil {
		s.log.Error("Update deployment image error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}
