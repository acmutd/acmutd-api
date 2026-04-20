package types

import "time"

type APIKey struct {
	// Core auth fields — used by the existing Auth/RateLimit middleware.
	// firestore tags are snake_case (storage format); json tags are camelCase (API format).
	Key           string    `firestore:"key" json:"key"`
	RateLimit     int       `firestore:"rate_limit" json:"rateLimit"`
	WindowSeconds int       `firestore:"window_seconds" json:"windowSeconds"`
	IsAdmin       bool      `firestore:"is_admin" json:"isAdmin"`
	CreatedAt     time.Time `firestore:"created_at" json:"createdAt"`
	ExpiresAt     time.Time `firestore:"expires_at" json:"expiresAt"`
	UsageCount    int64     `firestore:"usage_count" json:"usageCount"`

	// Dashboard fields — zero-valued for legacy keys; populated for dashboard-managed keys
	KeyID       string     `firestore:"key_id" json:"keyId"`
	UserID      string     `firestore:"user_id" json:"userId"`
	OwnerEmail  string     `firestore:"owner_email" json:"ownerEmail"`
	Label       string     `firestore:"label" json:"label"`
	Description string     `firestore:"description" json:"description"`
	Status      string     `firestore:"status" json:"status"` // "active" | "inactive" | "pending" | "rejected"
	LastUsedAt  *time.Time `firestore:"last_used_at" json:"lastUsedAt"`
}
