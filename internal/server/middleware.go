package server

import (
	"context"
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

func (api *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}

		// Check cache first
		if apiKeyData, found := api.apiKeyCache.Get(key); found {
			keyData := apiKeyData.(*types.APIKey)
			if keyData.ExpiresAt.Before(time.Now()) {
				api.apiKeyCache.Delete(key)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
				return
			}
			c.Set("api_key", apiKeyData.(*types.APIKey))
			c.Next()
			return
		}

		// Cache miss - check Firestore
		apiKey, err := api.db.ValidateAPIKey(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
		if apiKey == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		// Cache the valid key
		api.apiKeyCache.Set(key, apiKey, cache.DefaultExpiration)
		c.Set("api_key", apiKey)
		c.Next()
	}
}

func (api *Server) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		keyData, exists := c.Get("api_key")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "please provide an API key"})
			return
		}

		apiKey := keyData.(*types.APIKey)
		allowed := api.rateLimiter.Allow(apiKey.Key, apiKey.RateLimit, apiKey.WindowSeconds)

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}

func (api *Server) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}

		// check if the provided api key is the admin key
		if key != api.adminKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		go func(key string) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			api.db.UpdateKeyUsage(ctx, key)
		}(key)

		c.Next()
	}
}
