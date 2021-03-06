package src

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-api/types"
	"google.golang.org/grpc/status"
)

func (s *Server) postDeploymentHandler(c *gin.Context) {
	deploymentMsg := &types.DeploymentMessage{}
	bound := s.bindJSON(c, deploymentMsg)
	if bound != true {
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

func (s *Server) getDeploymentHandler(c *gin.Context) {
	depIDStr := c.Param("id")
	depID, err := strconv.ParseInt(depIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "deployment id is not int"})
		return
	}
	dep, err := s.cdClient.GetDeployment(depID)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok == true {
			if statusErr.Code() == http.StatusNotFound {
				c.JSON(http.StatusNotFound, gin.H{"err": "deployment not found"})
				return
			}
		}
		s.log.Error("Find deployment error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, dep)
}

func (s *Server) getAllDeploymentsHandler(c *gin.Context) {
	params := &types.PaginationQueryParams{
		Page:   1,
		Limit:  20,
	}
	err := c.ShouldBind(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
	}
	deps, err := s.cdClient.GetAllDeployments(*params)
	if err != nil {
		s.log.Error("Find deployments error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	deploymentArrayMsg, err := types.DeploymentArrayMsgFromProto(deps)
	if err != nil {
		s.log.Error("Failed to make deployment array message", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, deploymentArrayMsg)
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

func (s *Server) scaleDeploymentHandler(c *gin.Context) {
	depIDStr := c.Param("id")
	depID, err := strconv.ParseInt(depIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "deployment id is not int"})
		return
	}
	scaleMsg := &types.ScaleMessage{}
	bound := s.bindJSON(c, scaleMsg)
	if bound != true {
		return
	}
	if scaleMsg.Replicas <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "replicas count must be more than 0"})
		return
	}
	rpcMsg := &types.DeploymentMessage{
		ID:       depID,
		Replicas: scaleMsg.Replicas,
	}
	err = s.cdClient.ScaleDeployment(rpcMsg)
	if err != nil {
		s.log.Error("Scale deployment error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) updateManifestHandler(c *gin.Context) {
	depIDStr := c.Param("id")
	depID, err := strconv.ParseInt(depIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "deployment id is not int"})
		return
	}
	manifestMsg := &types.ManifestMessage{}
	bound := s.bindJSON(c, manifestMsg)
	if bound != true {
		return
	}
	if manifestMsg.Manifest == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "empty manifest"})
		return
	}
	rpcMsg := &types.DeploymentMessage{
		ID:       depID,
		Manifest: manifestMsg.Manifest,
	}
	err = s.cdClient.UpdateManifest(rpcMsg)
	if err != nil {
		s.log.Error("Update deployment manifest error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}
