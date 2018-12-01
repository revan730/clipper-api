package src

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-api/types"
)

func (s *Server) postBranchConfigHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	user, err := s.databaseClient.FindUserByID(userClaim.ID)
	if err != nil {
		s.logError("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		return
	}
	// TODO: Probably this part must be overwritten as pg
	// will return error on attemt to create config with non-existing
	// repo id
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if user.IsAdmin == false {
		if userClaim.ID != repo.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
			return
		}
	}
	branchMsg := &types.BranchMessage{}
	bound := s.bindJSON(c, branchMsg)
	if bound == false {
		return
	}
	if branchMsg.BranchName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "branch name not provided"})
		return
	}
	// Check if config for this branch already exists for this repo
	conf, err := s.databaseClient.FindBranchConfig(int64(repoID), branchMsg.BranchName)
	if err != nil {
		s.logError("Find branch config error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conf != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Config for this branch already exists"})
		return
	}
	branchConf := types.BranchConfig{
		GithubRepoID: int64(repoID),
		Branch:       branchMsg.BranchName,
		IsCiEnabled:  true,
	}
	err = s.databaseClient.CreateBranchConfig(&branchConf)
	if err != nil {
		s.logError("Create branch config error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) getAllBranchConfigsHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	// Check if user owns this repo
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if userClaim.IsAdmin == false {
		if userClaim.ID != repo.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
			return
		}
	}
	configs, err := s.databaseClient.FindAllBranchConfigs(repo.ID, c.Request.URL.Query())
	if err != nil {
		s.logError("Find branch configs error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func (s *Server) deleteBranchConfigHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	branchName := c.Param("branch")
	// Check if user owns this repo
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.logError("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if userClaim.IsAdmin == false {
		if userClaim.ID != repo.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
			return
		}
	}
	err = s.databaseClient.DeleteBranchConfig(int64(repoID), branchName)
	// TODO: Explicitly handle missing branch config error
	if err != nil {
		s.logError("Delete branch config error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}
