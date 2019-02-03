package src

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-api/types"
	"google.golang.org/grpc/status"
)

func (s *Server) getBuildArtifactHandler(c *gin.Context) {
	buildIDStr := c.Param("id")
	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "build id is not int"})
		return
	}
	artifact, err := s.ciClient.GetBuildArtifact(int64(buildID))
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok == true {
			if statusErr.Code() == http.StatusNotFound {
				c.JSON(http.StatusNotFound, gin.H{"err": "build artifact not found"})
				return
			}
		}
		s.log.Error("Find build artifact error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, artifact)
}

func (s *Server) getAllArtifactsHandler(c *gin.Context) {
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	params := &types.BuildsQueryParams{
		Branch: "master",
		Page:   1,
		Limit:  20,
	}
	err = c.ShouldBind(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
	}
	artifacts, err := s.ciClient.GetAllArtifacts(int64(repoID), *params)
	if err != nil {
		s.log.Error("Find artifacts error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	artifactArrayMsg := types.ArtifactArrayMsgFromProto(artifacts)
	c.JSON(http.StatusOK, artifactArrayMsg)
}
