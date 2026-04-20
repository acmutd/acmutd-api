package firebase

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const appConfigDocID = "main"

// GetAppConfig retrieves the singleton app configuration document.
// Returns nil (no error) if the document has not been created yet.
func (c *Firestore) GetAppConfig(ctx context.Context) (*types.AppConfig, error) {
	doc, err := c.Collection("app_config").Doc(appConfigDocID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	var cfg types.AppConfig
	return &cfg, doc.DataTo(&cfg)
}

// UpdateAppConfig applies a partial update to the singleton app config and returns the result.
// The caller provides a map of Firestore field paths to new values; updated_at is set automatically.
func (c *Firestore) UpdateAppConfig(ctx context.Context, updates map[string]any) (*types.AppConfig, error) {
	updates["updated_at"] = time.Now()

	ref := c.Collection("app_config").Doc(appConfigDocID)
	if _, err := ref.Set(ctx, updates, firestore.MergeAll); err != nil {
		return nil, err
	}

	return c.GetAppConfig(ctx)
}

// GetInstanceState retrieves the stored instance state from app_config/main.
// Returns a zero-value InstanceState if the document does not exist.
func (c *Firestore) GetInstanceState(ctx context.Context) (*types.InstanceState, error) {
	doc, err := c.Collection("app_config").Doc(appConfigDocID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return &types.InstanceState{State: "stopped"}, nil
		}
		return nil, err
	}

	data := doc.Data()
	state := &types.InstanceState{}

	if v, ok := data["instance_id"].(string); ok {
		state.InstanceID = v
	}
	if v, ok := data["instance_type"].(string); ok {
		state.InstanceType = v
	}
	if v, ok := data["instance_state"].(string); ok {
		state.State = v
	} else {
		state.State = "stopped"
	}
	if v, ok := data["public_ip"].(string); ok {
		state.PublicIP = v
	}
	if v, ok := data["uptime_seconds"].(int64); ok {
		state.UptimeSeconds = v
	}

	return state, nil
}

// UpdateInstanceState persists instance state fields into app_config/main.
func (c *Firestore) UpdateInstanceState(ctx context.Context, instanceState types.InstanceState) error {
	updates := map[string]any{
		"instance_id":    instanceState.InstanceID,
		"instance_type":  instanceState.InstanceType,
		"instance_state": instanceState.State,
		"public_ip":      instanceState.PublicIP,
		"uptime_seconds": instanceState.UptimeSeconds,
	}
	_, err := c.Collection("app_config").Doc(appConfigDocID).Set(ctx, updates, firestore.MergeAll)
	return err
}
