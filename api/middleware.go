package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/acmutd/acmutd-api/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (api *API) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || strings.HasPrefix(c.Request.URL.Path, "/admin") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required. Include 'Authorization: Bearer YOUR_API_KEY' header",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use 'Bearer YOUR_API_KEY'",
			})
			c.Abort()
			return
		}

		apiKey := parts[1]

		keyData, err := api.db.GetAPIKeyByKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		if !keyData.IsValid() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key is inactive or expired",
			})
			c.Abort()
			return
		}

		if !keyData.CanMakeRequest() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"rate_limit_info": types.RateLimitInfo{
					RemainingRequests: 0,
					ResetTime:         keyData.GetResetTime(),
					RateLimit:         keyData.RateLimit,
					RateInterval:      keyData.RateInterval,
				},
			})
			c.Abort()
			return
		}

		keyData.UpdateRateLimitWindow()

		if err := api.db.UpdateAPIKey(c.Request.Context(), keyData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update API key usage",
			})
			c.Abort()
			return
		}

		c.Set("api_key", keyData)

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", keyData.RateLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", keyData.GetRemainingRequests()))
		c.Header("X-RateLimit-Reset", keyData.GetResetTime().Format(time.RFC3339))

		c.Next()
	}
}

// GenerateAPIKey generates a secure random API key
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateAPIKeyID() (string, error) {
	newUUID := uuid.New()
	return newUUID.String(), nil
}

func (api *API) AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminToken := c.GetHeader("X-Admin-Token")
		if adminToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Admin token required",
			})
			c.Abort()
			return
		}

		expectedToken := os.Getenv("ADMIN_TOKEN")
		if expectedToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Admin token not configured",
			})
			c.Abort()
			return
		}
		if adminToken != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid admin token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
