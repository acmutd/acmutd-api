package firebase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func generateKeyID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "key_" + hex.EncodeToString(b)
}

func generateKeyValue() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// findKeyDocByKeyID returns the document reference and data for a key by its key_id field.
func (c *Firestore) findKeyDocByKeyID(ctx context.Context, keyID string) (*firestore.DocumentSnapshot, error) {
	iter := c.Collection("api_keys").Where("key_id", "==", keyID).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	return doc, err
}

// GetUserAPIKey returns the single API key owned by a given user, or nil if none exists.
func (c *Firestore) GetUserAPIKey(ctx context.Context, userID string) (*types.APIKey, error) {
	iter := c.Collection("api_keys").Where("user_id", "==", userID).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var key types.APIKey
	return &key, doc.DataTo(&key)
}

// ListAllKeys returns all dashboard-managed API keys ordered by created_at descending.
func (c *Firestore) ListAllKeys(ctx context.Context) ([]types.APIKey, error) {
	iter := c.Collection("api_keys").
		Where("key_id", "!=", "").
		OrderBy("key_id", firestore.Asc).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var keys []types.APIKey
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var k types.APIKey
		if err := doc.DataTo(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// ListPendingKeys returns all API keys with status "pending".
func (c *Firestore) ListPendingKeys(ctx context.Context) ([]types.APIKey, error) {
	iter := c.Collection("api_keys").
		Where("status", "==", "pending").
		OrderBy("created_at", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	var keys []types.APIKey
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var k types.APIKey
		if err := doc.DataTo(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// RequestAPIKey creates a new pending API key for a user.
// Only one key per user is allowed; returns an error if one already exists.
// The ExpiresAt date is taken from the app config's keys_expires_at_date field if set.
func (c *Firestore) RequestAPIKey(ctx context.Context, userID, ownerEmail, label, description string) error {
	existing, err := c.GetUserAPIKey(ctx, userID)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("user already has an API key")
	}

	cfg, err := c.GetAppConfig(ctx)
	if err != nil {
		return err
	}

	keyValue, err := generateKeyValue()
	if err != nil {
		return err
	}

	status := "pending"
	if cfg != nil {
		mode := cfg.AutoApproveMode
		switch mode {
		case "all":
			status = "active"
		case "@utdallas.edu", "@acmutd.co":
			if strings.HasSuffix(ownerEmail, mode) {
				status = "active"
			}
		case "both":
			if strings.HasSuffix(ownerEmail, "@utdallas.edu") || strings.HasSuffix(ownerEmail, "@acmutd.co") {
				status = "active"
			}
		}
	}

	key := types.APIKey{
		Key:           keyValue,
		KeyID:         generateKeyID(),
		UserID:        userID,
		OwnerEmail:    ownerEmail,
		Label:         label,
		Description:   description,
		Status:        status,
		IsAdmin:       false,
		RateLimit:     60,
		WindowSeconds: 60,
		CreatedAt:     time.Now(),
		UsageCount:    0,
	}
	if cfg != nil && !cfg.KeysExpiresAtDate.IsZero() {
		key.ExpiresAt = cfg.KeysExpiresAtDate
	}

	_, err = c.Collection("api_keys").Doc(keyValue).Set(ctx, key)
	return err
}

// setKeyStatus finds a key by key_id and updates its status field.
func (c *Firestore) setKeyStatus(ctx context.Context, keyID, newStatus string) error {
	doc, err := c.findKeyDocByKeyID(ctx, keyID)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("key not found: %s", keyID)
	}
	_, err = doc.Ref.Update(ctx, []firestore.Update{{Path: "status", Value: newStatus}})
	return err
}

func (c *Firestore) ApproveKey(ctx context.Context, keyID string) error {
	return c.setKeyStatus(ctx, keyID, "active")
}

func (c *Firestore) RejectKey(ctx context.Context, keyID string) error {
	return c.setKeyStatus(ctx, keyID, "rejected")
}

func (c *Firestore) RevokeKey(ctx context.Context, keyID string) error {
	return c.setKeyStatus(ctx, keyID, "inactive")
}

func (c *Firestore) DeleteKey(ctx context.Context, keyID string) error {
	doc, err := c.findKeyDocByKeyID(ctx, keyID)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("key not found: %s", keyID)
	}
	_, err = doc.Ref.Delete(ctx)
	return err
}

// RegenerateKey deletes the existing key document and creates a new one with fresh key bytes,
// preserving all metadata (key_id, user_id, owner_email, label, description, status).
func (c *Firestore) RegenerateKey(ctx context.Context, keyID string) (*types.APIKey, error) {
	doc, err := c.findKeyDocByKeyID(ctx, keyID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	var existing types.APIKey
	if err := doc.DataTo(&existing); err != nil {
		return nil, err
	}

	newKeyValue, err := generateKeyValue()
	if err != nil {
		return nil, err
	}

	updated := existing
	updated.Key = newKeyValue
	updated.UsageCount = 0
	updated.LastUsedAt = nil
	if existing.Status == "inactive" {
		updated.Status = "pending"
	}

	batch := c.BulkWriter(ctx)
	defer batch.End()
	batch.Delete(doc.Ref)
	batch.Set(c.Collection("api_keys").Doc(newKeyValue), updated)
	batch.Flush()

	return &updated, nil
}

// AddKey creates an API key directly with all fields specified (admin operation).
func (c *Firestore) AddKey(ctx context.Context, input types.APIKey) (*types.APIKey, error) {
	keyValue, err := generateKeyValue()
	if err != nil {
		return nil, err
	}

	input.Key = keyValue
	if input.KeyID == "" {
		input.KeyID = generateKeyID()
	}
	input.CreatedAt = time.Now()
	input.UsageCount = 0

	if _, err := c.Collection("api_keys").Doc(keyValue).Set(ctx, input); err != nil {
		return nil, err
	}
	return &input, nil
}

// UpdateKey updates mutable fields on a key identified by key_id.
func (c *Firestore) UpdateKey(ctx context.Context, keyID string, updates map[string]any) (*types.APIKey, error) {
	doc, err := c.findKeyDocByKeyID(ctx, keyID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	firestoreUpdates := make([]firestore.Update, 0, len(updates))
	for path, val := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: path, Value: val})
	}

	if _, err := doc.Ref.Update(ctx, firestoreUpdates); err != nil {
		return nil, err
	}

	refreshed, err := doc.Ref.Get(ctx)
	if err != nil {
		return nil, err
	}
	var key types.APIKey
	return &key, refreshed.DataTo(&key)
}

// GetKeyByKeyID retrieves a key by its key_id field.
func (c *Firestore) GetKeyByKeyID(ctx context.Context, keyID string) (*types.APIKey, error) {
	doc, err := c.findKeyDocByKeyID(ctx, keyID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	var key types.APIKey
	return &key, doc.DataTo(&key)
}
