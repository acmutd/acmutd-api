package types

import "time"

type APIKey struct {
	Key           string    `firestore:"key" json:"key"`
	RateLimit     int       `firestore:"rate_limit" json:"rate_limit"`         // Maximum requests allowed per window
	WindowSeconds int       `firestore:"window_seconds" json:"window_seconds"` // Time window in seconds for rate limiting
	IsAdmin       bool      `firestore:"is_admin" json:"is_admin"`             // Whether the key has admin privileges (no rate limiting)
	CreatedAt     time.Time `firestore:"created_at" json:"created_at"`
	ExpiresAt     time.Time `firestore:"expires_at" json:"expires_at"`   // Expiration date for the key
	UsageCount    int64     `firestore:"usage_count" json:"usage_count"` // Number of times the key has been used
}
