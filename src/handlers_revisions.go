package src

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-api/types"
)

func (s *Server) getRevisionsHandler(c *gin.Context) {
	deploymentIDStr := c.Param("id")
	deploymentID, err := strconv.ParseInt(deploymentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	params := &types.PaginationQueryParams{
		Page:   1,
		Limit:  20,
	}
	err = c.ShouldBind(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
	}
	revisions, err := s.cdClient.GetRevisions(deploymentID, *params)
	if err != nil {
		s.log.Error("Find revisions error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	revisionArrayMsg, err := types.RevisionArrayMsgFromProto(revisions)
	if err != nil {
		s.log.Error("Failed to make revision array message", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, revisionArrayMsg)
}