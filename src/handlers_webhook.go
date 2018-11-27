package src

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/revan730/clipper-api/types"
	commonTypes "github.com/revan730/clipper-common/types"
)

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

func checkSecret(secret string, c *gin.Context) error {
	if secret == "" {
		return nil
	}
	gitSignStr := c.GetHeader("X-Hub-Signature")
	if gitSignStr == "" {
		return errors.New("Github signature not provided")
	}
	rawMsg, ok := c.Get(gin.BodyBytesKey)
	if ok != true {
		return errors.New("Failed to get request body")
	}
	body, ok := rawMsg.([]byte)
	if ok != true {
		return errors.New("Failed to assert request body")
	}
	actualSign := make([]byte, 20)
	hex.Decode(actualSign, []byte(gitSignStr[5:]))

	if hmac.Equal(signBody([]byte(secret), body), actualSign) == false {
		return errors.New("Signature doesn't match")
	}
	return nil
}

func (s *Server) webhookHandler(c *gin.Context) {
	payload := &types.WebhookMessage{}
	err := c.ShouldBindBodyWith(&payload, binding.JSON)
	if err != nil {
		s.logError("JSON read error", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(payload)
	// Get user by user param
	userLogin := c.Param("user")
	user, err := s.databaseClient.FindUser(userLogin)
	if err != nil {
		s.logError("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = checkSecret(user.WebhookSecret, c)
	if err != nil {
		s.logError("Webhook secret error", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	eventType := c.GetHeader("X-GitHub-Event")
	switch eventType {
	case "push":
		repo, err := s.databaseClient.FindRepoByName(payload.Repository.FullName)
		if err != nil {
			s.logError("Failed to find repo", err)
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		branchName := strings.Split(payload.Ref, "/")[2]
		config, err := s.databaseClient.FindBranchConfig(repo.ID, branchName)
		if err != nil {
			s.logError("Failed to find branch config", err)
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Automatic CI is not enabled for this branch, ignore
		if config == nil || config.IsCiEnabled == false {
			c.Writer.WriteHeader(http.StatusOK)
			return
		}
		ciMsg := commonTypes.CIJob{
			RepoURL:     payload.Repository.GitURL,
			Branch:      branchName,
			HeadSHA:     payload.HeadCommit.SHA,
			User:        user.Login,
			AccessToken: user.AccessToken,
			RepoID:      repo.ID,
		}
		err = s.startCIJob(ciMsg)
		if err != nil {
			s.logError("Failed to publish CI job", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		return
	case "pull_request":
		// Here branch name can be found in
		// payload.Head.Ref
		if payload.Action != "opened" && payload.Action != "synchronize" {
			// We can only know by this event that
			// ci is required if pull request opens or syncronizes
			c.Writer.WriteHeader(http.StatusOK)
			return
		}
		repo, err := s.databaseClient.FindRepoByName(payload.Repository.FullName)
		if err != nil {
			s.logError("Failed to find repo", err)
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		ciMsg := commonTypes.CIJob{
			RepoURL:     payload.Repository.GitURL,
			Branch:      payload.PullRequest.Head.Ref,
			HeadSHA:     payload.PullRequest.Head.SHA,
			User:        user.Login,
			AccessToken: user.AccessToken,
			RepoID:      repo.ID,
		}
		err = s.startCIJob(ciMsg)
		if err != nil {
			s.logError("Failed to publish CI job", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		s.logInfo("Unsupported type, ignoring")
		c.Writer.WriteHeader(http.StatusOK)
		return
	}
}
