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
		s.logError("Find build error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	buildMsg, err := types.BuildMsgFromBuild(build)
	if err != nil {
		s.logError("Failed to make build message", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, buildMsg)
}
