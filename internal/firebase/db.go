package firebase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/patrickmn/go-cache"
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
	*cache.Cache
	TTL FirestoreConfig
}

type FirestoreConfig struct {
	Sections   time.Duration
	Courses    time.Duration
	Professors time.Duration
	Terms      time.Duration
}

// NewFirestore creates a new Firestore client from a Firebase app
func NewFirestore(ctx context.Context, app *firebase.App) (*Firestore, error) {
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %w", err)
	}

	return &Firestore{
		Client: client,
		Cache:  cache.New(5*time.Minute, 10*time.Minute),
		TTL: FirestoreConfig{
			Sections:   30 * time.Minute,
			Courses:    2 * time.Hour,
			Professors: 2 * time.Hour,
			Terms:      6 * time.Hour,
		},
	}, nil
}

// cachedResult holds a query result alongside its pagination state for caching
type cachedResult[T any] struct {
	Items   []T
	HasNext bool
}

// cacheKey builds a deterministic cache key from a collection name and alternating
// key-value pairs.
//   - Empty values are omitted.
//   - Pairs are sorted so argument order does not affect the result.
//
// Example:
//
//	cacheKey("sections", "term", "25s", "prefix", "cs") → "sections::prefix=cs::term=25s"
func cacheKey(collection string, kvs ...string) string {
	pairs := make([]string, 0, len(kvs)/2)
	for i := 0; i+1 < len(kvs); i += 2 {
		if kvs[i+1] != "" {
			pairs = append(pairs, kvs[i]+"="+kvs[i+1])
		}
	}
	sort.Strings(pairs)
	if len(pairs) == 0 {
		return collection
	}
	return collection + "::" + strings.Join(pairs, "::")
}
