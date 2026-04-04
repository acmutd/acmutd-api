package firebase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

// ResolveConfigFilename returns the Firebase service account JSON path
// based on FB_CONFIG env var and the target environment (dev or prod).
func ResolveConfigFilename(env string) (string, error) {
	baseName := strings.TrimSpace(os.Getenv("FB_CONFIG"))
	if baseName == "" {
		return "", errors.New("FB_CONFIG is required")
	}

	switch env {
	case "prod":
		return "prod." + baseName, nil
	case "dev", "local":
		return "dev." + baseName, nil
	default:
		return "", errors.New("invalid environment")
	}
}

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
