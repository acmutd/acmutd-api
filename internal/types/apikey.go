package types

import "time"

type APIKey struct {
	Key           string    `firestore:"key" json:"key"`
	RateLimit     int       `firestore:"rate_limit" json:"rate_limit"`
	WindowSeconds int       `firestore:"window_seconds" json:"window_seconds"`
	IsAdmin       bool      `firestore:"is_admin" json:"is_admin"`
	CreatedAt     time.Time `firestore:"created_at" json:"created_at"`
	ExpiresAt     time.Time `firestore:"expires_at" json:"expires_at"`
	LastUsedAt    time.Time `firestore:"last_used_at" json:"last_used_at"`
	UsageCount    int64     `firestore:"usage_count" json:"usage_count"`
}
