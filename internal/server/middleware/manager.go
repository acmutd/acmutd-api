package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/acmutd/acmutd-api/internal/server/ratelimit"
	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// Manager wires all HTTP middlewares with shared dependencies.
type Manager struct {
	db          *firebase.Firestore
	apiKeyCache *cache.Cache
	rateLimiter *ratelimit.Limiter
	adminKey    string
}

// NewManager builds a middleware manager for the HTTP server.
func NewManager(db *firebase.Firestore, apiKeyCache *cache.Cache, limiter *ratelimit.Limiter, adminKey string) *Manager {
	return &Manager{
		db:          db,
		apiKeyCache: apiKeyCache,
		rateLimiter: limiter,
		adminKey:    adminKey,
	}
}

// Auth validates API keys and decorates the context with key metadata.
func (m *Manager) Auth() gin.HandlerFunc {
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

		if apiKeyData, found := m.apiKeyCache.Get(key); found {
			keyData, ok := apiKeyData.(*types.APIKey)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
				return
			}

			if keyData.IsAdmin {
				m.updateKeyUsageAsync(key)
				c.Set("api_key", keyData)
				c.Next()
				return
			}

			if keyData.ExpiresAt.Before(time.Now()) {
				m.apiKeyCache.Delete(key)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
				return
			}

			m.updateKeyUsageAsync(key)

			c.Set("api_key", keyData)
			c.Next()
			return
		}

		apiKey, err := m.db.ValidateAPIKey(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
		if apiKey == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		if apiKey.ExpiresAt.Before(time.Now()) && !apiKey.IsAdmin {
			m.apiKeyCache.Delete(key)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
			return
		}

		m.updateKeyUsageAsync(key)
		m.apiKeyCache.Set(key, apiKey, cache.DefaultExpiration)
		c.Set("api_key", apiKey)
		c.Next()
	}
}

// RateLimit enforces per-key request limits.
func (m *Manager) RateLimit() gin.HandlerFunc {
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
		if !m.rateLimiter.Allow(apiKey.Key, apiKey.RateLimit, apiKey.WindowSeconds) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}

// Admin restricts routes to the generated admin key.
func (m *Manager) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}

		if key != m.adminKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		m.updateKeyUsageAsync(key)
		c.Next()
	}
}

func (m *Manager) updateKeyUsageAsync(key string) {
	if key == "" {
		return
	}

	go func(k string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		m.db.UpdateKeyUsage(ctx, k)
	}(key)
}
