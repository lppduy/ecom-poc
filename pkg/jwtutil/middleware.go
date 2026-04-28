package jwtutil

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ContextKeyUserID = "userID"

// RequireAuth extracts and validates JWT from Authorization: Bearer <token>.
// Sets userID in gin context on success.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := Verify(tokenStr, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set(ContextKeyUserID, claims.UserID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from gin context.
func GetUserID(c *gin.Context) string {
	id, _ := c.Get(ContextKeyUserID)
	if s, ok := id.(string); ok {
		return s
	}
	return ""
}
