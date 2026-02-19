package firebase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
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
		return nil, err
	}

	var apiKey types.APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// DeleteAllAdminKeys deletes all existing admin keys from Firebase
func (c *Firestore) DeleteAllAdminKeys(ctx context.Context) (returnedErr error) {
	// Query all documents in api_keys collection where is_admin is true
	iter := c.Collection("api_keys").Where("is_admin", "==", true).Documents(ctx)
	defer iter.Stop()

	batch := c.BulkWriter(ctx)
	defer batch.End()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if status.Code(err) == codes.NotFound {
				break
			}
			return fmt.Errorf("failed to iterate admin keys: %w", err)
		}

		batch.Delete(doc.Ref)
	}

	return nil
}

// GenerateAdminAPIKey generates a new admin API key with the "admin-" prefix
func (c *Firestore) GenerateAdminAPIKey(ctx context.Context) (string, error) {
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	baseKey := hex.EncodeToString(keyBytes)
	adminKey := "admin-" + baseKey

	apiKey := types.APIKey{
		Key:           adminKey,
		RateLimit:     0, // No rate limit for admin
		WindowSeconds: 0,
		IsAdmin:       true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Time{}, // Never expires
		UsageCount:    0,
	}

	// Store with the full admin key as the document ID
	_, err := c.Collection("api_keys").Doc(adminKey).Set(ctx, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to store admin key: %w", err)
	}

	return adminKey, nil
}
