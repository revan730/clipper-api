package src

import (
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/revan730/clipper-api/types"
)

func jwtMiddleware(secret []byte) gin.HandlerFunc {
	// TODO: Json error handler
	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	return func(c *gin.Context) {
		err := jwtMiddleware.CheckJWT(c.Writer, c.Request)
		if err != nil {
			return
		}
		c.Next()
	}
}

func (s *Server) getClaimByName(c *gin.Context, name string) interface{} {
	jwtToken := c.Request.Context().Value("user")
	claims, ok := jwtToken.(*jwt.Token).Claims.(jwt.MapClaims)
	if ok == false {
		return nil
	}
	return claims[name]
}

func (s *Server) getUserLoginClaim(c *gin.Context) (string, bool) {
	login, ok := s.getClaimByName(c, "user").(string)
	if ok == false {
		s.log.Info("Failed to get user login claim")
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return "", false
	}
	return login, true
}

func (s *Server) getUserIDClaim(c *gin.Context) (int64, bool) {
	userID, ok := s.getClaimByName(c, "userID").(float64)
	if ok == false {
		s.log.Info("Failed to get userID claim")
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return 0, false
	}
	return int64(userID), true
}

func (s *Server) getUserAdminClaim(c *gin.Context) (bool, bool) {
	isAdmin, ok := s.getClaimByName(c, "admin").(bool)
	if ok == false {
		s.log.Info("Failed to get user admin claim")
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return false, false
	}
	return isAdmin, true
}

func (s *Server) userClaimMiddleware(c *gin.Context) {
	userID, ok := s.getUserIDClaim(c)
	if ok == false {
		return
	}
	login, ok := s.getUserLoginClaim(c)
	if ok == false {
		return
	}
	isAdmin, ok := s.getUserAdminClaim(c)
	if ok == false {
		return
	}
	user := types.User{
		ID:      userID,
		IsAdmin: isAdmin,
		Login:   login,
	}
	c.Set("userClaim", user)
	c.Next()
}
