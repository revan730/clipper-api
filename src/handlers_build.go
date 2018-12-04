package src

import (
	"errors"
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
	buildMsg, err := types.BuildMsgFromProto(build)
	if err != nil {
		s.logError("Failed to make build message", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, buildMsg)
}

func getBuidsQueryParams(queryParams map[string][]string) (*types.BuildsQueryParams, error) {
	params := &types.BuildsQueryParams{
		Page:  1,
		Limit: 20,
	}
	if len(queryParams["branch"]) > 0 {
		params.Branch = queryParams["branch"][0]
	}
	if params.Branch == "" {
		params.Branch = "master"
	}
	if len(queryParams["page"]) > 0 {
		pageStr := queryParams["page"][0]
		if pageStr != "" {
			page, err := strconv.Atoi(pageStr)
			if err != nil {
				return nil, errors.New("page is not int")
			}
			params.Page = page
		}
	}
	if len(queryParams["limit"]) > 0 {
		limitStr := queryParams["limit"][0]
		if limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				return nil, errors.New("limit is not int")
			}
			params.Limit = limit
		}
	}
	return params, nil
}

func (s *Server) getAllBuildsHandler(c *gin.Context) {
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	queryParams := c.Request.URL.Query()
	parsedParams, err := getBuidsQueryParams(queryParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
	}
	builds, err := s.ciClient.GetAllBuilds(int64(repoID), *parsedParams)
	if err != nil {
		s.logError("Find build error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	buildArrayMsg, err := types.BuildArrayMsgFromProto(builds)
	if err != nil {
		s.logError("Failed to make build array message", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, buildArrayMsg)
}
