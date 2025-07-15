package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a custom type that can unmarshal from JSON strings
type Duration time.Duration

// Need this to unmarshal the rate_interval field from the API key request
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s", s)
	}

	*d = Duration(duration)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

type APIKey struct {
	ID               string        `json:"id" firestore:"id"`
	Key              string        `json:"key" firestore:"key"`
	ExpiresAt        time.Time     `json:"expires_at" firestore:"expires_at"`
	LastUsedAt       time.Time     `json:"last_used_at" firestore:"last_used_at"`
	UsageCount       int64         `json:"usage_count" firestore:"usage_count"`
	RateLimit        int           `json:"rate_limit" firestore:"rate_limit"`
	RateInterval     time.Duration `json:"rate_interval" firestore:"rate_interval"`
	CurrentWindow    time.Time     `json:"current_window" firestore:"current_window"`
	RequestsInWindow int           `json:"requests_in_window" firestore:"requests_in_window"`
	IsActive         bool          `json:"is_active" firestore:"is_active"`
	CreatedAt        time.Time     `json:"created_at" firestore:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" firestore:"updated_at"`
}

type APIKeyRequest struct {
	RateLimit    int        `json:"rate_limit" binding:"required,min=1,max=1000"`
	RateInterval Duration   `json:"rate_interval" binding:"required"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

// Validate checks if the APIKeyRequest is valid
func (req *APIKeyRequest) Validate() error {
	if req.RateLimit < 1 || req.RateLimit > 1000 {
		return fmt.Errorf("rate_limit must be between 1 and 1000")
	}

	duration := req.RateInterval.ToDuration()
	if duration < time.Minute || duration > 24*time.Hour {
		return fmt.Errorf("rate_interval must be between 1m and 24h")
	}

	return nil
}

type APIKeyResponse struct {
	ID           string    `json:"id"`
	Key          string    `json:"key"`
	ExpiresAt    time.Time `json:"expires_at"`
	RateLimit    int       `json:"rate_limit"`
	RateInterval string    `json:"rate_interval"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UsageCount   int64     `json:"usage_count"`
	LastUsedAt   time.Time `json:"last_used_at"`
}

type RateLimitInfo struct {
	RemainingRequests int           `json:"remaining_requests"`
	ResetTime         time.Time     `json:"reset_time"`
	RateLimit         int           `json:"rate_limit"`
	RateInterval      time.Duration `json:"rate_interval"`
}

func (ak *APIKey) IsExpired() bool {
	return !ak.ExpiresAt.IsZero() && time.Now().After(ak.ExpiresAt)
}

func (ak *APIKey) IsValid() bool {
	return ak.IsActive && !ak.IsExpired()
}

func (ak *APIKey) UpdateRateLimitWindow() {
	now := time.Now()

	// Check if current window has expired
	if now.After(ak.CurrentWindow) {
		ak.CurrentWindow = ak.CurrentWindow.Add(ak.RateInterval)
		ak.RequestsInWindow = 1
	} else {
		ak.RequestsInWindow++
	}

	ak.LastUsedAt = now
	ak.UsageCount++
}

func (ak *APIKey) CanMakeRequest() bool {
	now := time.Now()

	if now.After(ak.CurrentWindow) {
		return true
	}

	return ak.RequestsInWindow < ak.RateLimit
}

func (ak *APIKey) GetRemainingRequests() int {
	now := time.Now()

	if now.After(ak.CurrentWindow) {
		return ak.RateLimit
	}

	remaining := ak.RateLimit - ak.RequestsInWindow
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (ak *APIKey) GetResetTime() time.Time {
	return ak.CurrentWindow
}
