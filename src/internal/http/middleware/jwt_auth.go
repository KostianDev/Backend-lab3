package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/storage"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	ContextUserID       = "userID"
	ContextEmail        = "email"
)

// JWTAuth creates middleware that validates JWT tokens from Authorization header.
func JWTAuth(jwtService *storage.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(AuthorizationHeader)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "missing authorization header",
				},
			})
			return
		}

		if !strings.HasPrefix(header, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "invalid authorization format",
				},
			})
			return
		}

		tokenString := strings.TrimPrefix(header, BearerPrefix)
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": err.Error(),
				},
			})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from context.
func GetUserID(c *gin.Context) (uint, bool) {
	id, exists := c.Get(ContextUserID)
	if !exists {
		return 0, false
	}
	userID, ok := id.(uint)
	return userID, ok
}

// GetEmail extracts the authenticated email from context.
func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(ContextEmail)
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}
