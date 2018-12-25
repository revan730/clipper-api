package src

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"github.com/revan730/clipper-api/types"
)

const repoRegexpStr = `^[a-zA-Z0-9-_]+\/[a-zA-Z0-9-_]+$`

var repoRegexp *regexp.Regexp

func validateRepoName(name string) bool {
	if repoRegexp == nil {
		repoRegexp, _ = regexp.Compile(repoRegexpStr)
	}
	return repoRegexp.MatchString(name)
}

func (s *Server) postRepoHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoMsg := &types.RepoMessage{}
	bound := s.bindJSON(c, repoMsg)
	if bound == false {
		return
	}
	if repoMsg.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo name not provided"})
		return
	}
	ok := validateRepoName(repoMsg.FullName)
	if ok != true {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo name is not valid"})
		return
	}
	err := s.databaseClient.CreateRepo(repoMsg.FullName, userClaim.ID)
	if err != nil {
		// TODO: Maybe move this error handling to CreateUser func?
		pgErr, ok := err.(pg.Error)
		if ok && pgErr.IntegrityViolation() {
			c.JSON(http.StatusBadRequest, gin.H{"err": "Repo already exists"})
			return
		}
		s.log.Error("Create repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) getRepoHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.log.Error("Find repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if repo == nil {
		c.JSON(http.StatusNotFound, gin.H{"err": "repo not found"})
		return
	}
	if repo.UserID != userClaim.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "you have no access to this repo"})
	}
	c.JSON(http.StatusOK, repo)
}

func (s *Server) getAllReposHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repos, err := s.databaseClient.FindAllUserRepos(userClaim.ID, c.Request.URL.Query())
	if err != nil {
		s.log.Error("Find repos error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"repos": repos})
}

func (s *Server) deleteRepoHandler(c *gin.Context) {
	userClaim := c.MustGet("userClaim").(types.User)
	repoIDStr := c.Param("id")
	repoID, err := strconv.Atoi(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "repo id is not int"})
		return
	}
	// TODO: Don't think this part is effective, as user who
	// forges access tokens can simply know user's id from
	// token itself, no need to guess
	user, err := s.databaseClient.FindUserByID(userClaim.ID)
	if err != nil {
		s.log.Error("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "no access"})
		return
	}
	repo, err := s.databaseClient.FindRepoByID(int64(repoID))
	if err != nil {
		s.log.Error("Find repo error", err)
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
	// TODO: Explicitly handle missing repo error
	err = s.databaseClient.DeleteRepoByID(int64(repoID))
	if err != nil {
		s.log.Error("Delete repo error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}
