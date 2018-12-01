package src

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
		s.logError("Find build error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if build == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "build not found"})
		return
	}
	// TODO: Replace timestamp with time
	c.JSON(http.StatusOK, build)
}
