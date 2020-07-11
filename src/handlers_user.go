package src

import (
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"github.com/revan730/clipper-api/types"
)

func (s *Server) loginHandler(c *gin.Context) {
	// Check if login and password are provided
	loginMsg := &types.CredentialsMessage{}
	bound := s.bindJSON(c, loginMsg)
	if bound != true {
		return
	}
	if loginMsg.Login == "" || loginMsg.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Empty login or password"})
		return
	}
	user, err := s.databaseClient.FindUser(loginMsg.Login)
	if err != nil {
		s.log.Error("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Failed to login"})
		return
	}
	if user.Authenticate(loginMsg.Password) == false {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "Failed to login"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin":  user.IsAdmin,
		"user":   user.Login,
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		s.log.Error("jwt error", err)
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (s *Server) registerHandler(c *gin.Context) {
	// Check if login and password are provided
	registerMsg := &types.CredentialsMessage{}
	bound := s.bindJSON(c, registerMsg)
	if bound != true {
		return
	}
	if registerMsg.Login == "" || registerMsg.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Empty login or password"})
		return
	}
	err := s.databaseClient.CreateUser(registerMsg.Login, registerMsg.Password, false)
	if err != nil {
		// TODO: Maybe move this error handling to CreateUser func?
		pgErr, ok := err.(pg.Error)
		if ok && pgErr.IntegrityViolation() {
			c.JSON(http.StatusBadRequest, gin.H{"err": "User already exists"})
			return
		}
		s.log.Error("Create user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) setSecretHandler(c *gin.Context) {
	secretMsg := &types.WebhookSecretMessage{}
	bound := s.bindJSON(c, secretMsg)
	if bound == false {
		return
	}
	if secretMsg.Secret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "secret not provided"})
		return
	}
	userClaim := c.MustGet("userClaim").(types.User)
	user, err := s.databaseClient.FindUser(userClaim.Login)
	if err != nil {
		s.log.Error("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "user not found"})
		return
	}
	user.WebhookSecret = secretMsg.Secret
	err = s.databaseClient.SaveUser(user)
	if err != nil {
		s.log.Error("User save error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) setAccessTokenHandler(c *gin.Context) {
	tokenMsg := &types.AccessTokenMessage{}
	bound := s.bindJSON(c, tokenMsg)
	if bound == false {
		return
	}
	if tokenMsg.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "access token not provided"})
		return
	}
	userClaim := c.MustGet("userClaim").(types.User)
	user, err := s.databaseClient.FindUser(userClaim.Login)
	if err != nil {
		s.log.Error("Find user error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"err": "user not found"})
		return
	}
	user.AccessToken = tokenMsg.Token
	err = s.databaseClient.SaveUser(user)
	if err != nil {
		s.log.Error("User save error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"err": nil})
}

func (s *Server) getAllUsersHandler(c *gin.Context) {
	users, err := s.databaseClient.FindAllUsers(c.Request.URL.Query())
	if err != nil {
		s.log.Error("Find all users error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	count, err := s.databaseClient.FindAllUsersCount()
	if err != nil {
		s.log.Error("Find all users count error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	usersMsg := &types.UsersArrayMessage{
		Total: count,
		Users: users,
	}

	c.JSON(http.StatusOK, usersMsg)
}

func (s *Server) changeUserAdminHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "user id is not int"})
		return
	}
	userAdminMsg := &types.UserAdminMessage{}
	bound := s.bindJSON(c, userAdminMsg)
	if bound != true {
		return
	}

	err = s.databaseClient.ChangeUserAdminStatus(userID, userAdminMsg.IsAdmin)
	if err != nil {
		s.log.Error("Change user admin status error", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"err": nil})
}
