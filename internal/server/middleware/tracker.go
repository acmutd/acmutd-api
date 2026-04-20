package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type responseWriter struct {
	gin.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// Track records a RequestEvent and upserts a DailyStat for every authenticated /api/v1 request.
// It is a no-op when trackStats is false. All Firestore writes are fire-and-forget.
func (m *Manager) Track() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: c.Writer}
		c.Writer = rw

		c.Next()

		if !m.trackStats.Load() {
			return
		}

		raw, exists := c.Get("api_key")
		if !exists {
			return
		}
		apiKey, ok := raw.(*types.APIKey)
		if !ok || apiKey == nil || apiKey.KeyID == "" {
			return
		}

		statusCode := rw.status
		if statusCode == 0 {
			statusCode = c.Writer.Status()
		}
		latencyMs := int(time.Since(start).Milliseconds())
		endpoint := c.FullPath()
		now := time.Now().UTC()
		success := statusCode >= 200 && statusCode < 400

		event := &types.RequestEvent{
			ID:         uuid.New().String(),
			UserID:     apiKey.UserID,
			KeyID:      apiKey.KeyID,
			DateTime:   now,
			Endpoint:   endpoint,
			Method:     c.Request.Method,
			StatusCode: statusCode,
			LatencyMs:  latencyMs,
		}

		date := now.Format("2006-01-02")

		go func(e *types.RequestEvent, kID, uID, d, ep string, suc bool) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()

			if err := m.db.RecordRequestEvent(ctx, e); err != nil {
				log.Printf("[tracker] failed to record request event: %v", err)
			}
			if err := m.db.IncrementDailyStat(ctx, kID, uID, d, ep, suc); err != nil {
				log.Printf("[tracker] failed to increment daily stat: %v", err)
			}
		}(event, apiKey.KeyID, apiKey.UserID, date, endpoint, success)
	}
}
