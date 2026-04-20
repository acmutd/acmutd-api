package firebase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Firestore) GenerateAPIKey(
	ctx context.Context,
	rateLimit int,
	windowSeconds int,
	isAdmin bool,
	expiresAt time.Time,
) (string, error) {
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	key := hex.EncodeToString(keyBytes)

	apiKey := types.APIKey{
		Key:           key,
		RateLimit:     rateLimit,
		WindowSeconds: windowSeconds,
		IsAdmin:       isAdmin,
		CreatedAt:     time.Now(),
		ExpiresAt:     expiresAt,
		UsageCount:    0,
	}

	_, err := c.Collection("api_keys").Doc(key).Set(ctx, apiKey)
	return key, err
}

// ValidateAPIKey with expiration check
func (c *Firestore) ValidateAPIKey(ctx context.Context, key string) (*types.APIKey, error) {
	doc, err := c.Collection("api_keys").Doc(key).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var apiKey types.APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, err
	}

	// We don't need to check expiration here because it's checked in the middleware
	return &apiKey, nil
}

// UpdateKeyUsage updates last used and usage count
func (c *Firestore) UpdateKeyUsage(ctx context.Context, key string) error {
	_, err := c.Collection("api_keys").Doc(key).Update(ctx, []firestore.Update{
		{Path: "usage_count", Value: firestore.Increment(1)},
	})
	return err
}

func (c *Firestore) GetAPIKey(ctx context.Context, key string) (*types.APIKey, error) {
	doc, err := c.Collection("api_keys").Doc(key).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var apiKey types.APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, err
	}

	return &apiKey, nil
}
