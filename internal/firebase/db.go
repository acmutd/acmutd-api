package firebase

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

// Firestore wraps the Firestore client and provides database operations
type Firestore struct {
	*firestore.Client
}

// NewFirestore creates a new Firestore client from a Firebase app
func NewFirestore(ctx context.Context, app *firebase.App) (*Firestore, error) {
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %w", err)
	}

	return &Firestore{
		Client: client,
	}, nil
}

// sanitizeDocID sanitizes a value for use as a Firestore document ID
func sanitizeDocID(value string) string {
	sanitized := strings.TrimSpace(value)
	sanitized = strings.ReplaceAll(sanitized, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "")
	return sanitized
}
