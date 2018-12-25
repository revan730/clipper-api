package src

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-api/types"
	"google.golang.org/grpc/status"
)

func (s *Server) getBuildHandler(c *gin.Context) {
	buildIDStr := c.Param("id")
	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "build id is not int"})
		return
	}
	build, err := s.ciClient.GetBuild(int64(buildID))
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok == true {
			if statusErr.Code() == http.StatusNotFound {
				c.JSON(http.StatusNotFound, gin.H{"err": "build not found"})
				return
			}
		}
		s.log.Error("Find build error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	buildMsg, err := types.BuildMsgFromProto(build)
	if err != nil {
		s.log.Error("Failed to make build message", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, buildMsg)
}

func (s *Server) getAllBuildsHandler(c *gin.Context) {
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
	builds, err := s.ciClient.GetAllBuilds(int64(repoID), *params)
	if err != nil {
		s.log.Error("Find build error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	buildArrayMsg, err := types.BuildArrayMsgFromProto(builds)
	if err != nil {
		s.log.Error("Failed to make build array message", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, buildArrayMsg)
}
