package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUID     = "dash_uid"
	ContextKeyIsAdmin = "dash_is_admin"
)

// FirebaseAuth validates a Firebase ID token from the Authorization: Bearer header.
// On success it upserts the user in Firestore (best-effort) and sets ContextKeyUID
// and ContextKeyIsAdmin in the gin.Context.
func (m *Manager) FirebaseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := m.authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		user, err := m.db.CreateOrUpdateUser(c.Request.Context(), token.UID, token.Claims)
		if err != nil {
			log.Printf("[dashboard] failed to upsert user %s: %v", token.UID, err)
		}

		isAdmin := false
		if user != nil {
			isAdmin = user.IsAdmin
		}

		c.Set(ContextKeyUID, token.UID)
		c.Set(ContextKeyIsAdmin, isAdmin)
		c.Next()
	}
}

// DashboardAdmin rejects requests from users whose is_admin flag is false.
// Must be used after FirebaseAuth.
func (m *Manager) DashboardAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, _ := c.Get(ContextKeyIsAdmin)
		if admin, ok := isAdmin.(bool); !ok || !admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
