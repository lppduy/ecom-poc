package jwtutil

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ContextKeyUserID = "userID"

type contextKey string

const contextKeyToken contextKey = "authToken"

// RequireAuth extracts and validates JWT from Authorization: Bearer <token>.
// Sets userID in gin context and stores raw token in request context for forwarding.
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
		// Store raw token in request context so downstream service clients can forward it
		ctx := context.WithValue(c.Request.Context(), contextKeyToken, tokenStr)
		c.Request = c.Request.WithContext(ctx)
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

// TokenFromContext retrieves the raw JWT token from a stdlib context.
// Used by HTTP clients to forward the token to downstream services.
func TokenFromContext(ctx context.Context) string {
	t, _ := ctx.Value(contextKeyToken).(string)
	return t
}
