package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	fb "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/patrickmn/go-cache"
	"google.golang.org/api/option"
)

const (
	apiKeyCacheTTL    = 5 * time.Minute
	rateLimitCacheTTL = 1 * time.Minute
)

type Server struct {
	db          *firebase.Firestore
	apiKeyCache *cache.Cache
	rateLimiter *RateLimiter
	port        int
	adminKey    string
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	sa := option.WithCredentialsFile(os.Getenv("FIREBASE_CONFIG"))
	app, err := fb.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	db, err := firebase.NewFirestore(context.Background(), app)
	if err != nil {
		log.Fatalf("error initializing firestore: %v\n", err)
	}

	// check for existing admin key by using reserved doc ID
	ctx := context.Background()
	var adminKey string
	const adminKeyDocID = "admin"
	adminKeyDoc, err := db.Client.Collection("api_keys").Doc(adminKeyDocID).Get(ctx)
	if err == nil && adminKeyDoc.Exists() {
		var adminKeyObj struct{ Key string }
		if err := adminKeyDoc.DataTo(&adminKeyObj); err == nil {
			adminKey = adminKeyObj.Key
		}
	}
	if adminKey == "" {
		// generate new admin key and add prefix
		key, err := db.GenerateAPIKey(ctx, 0, 0, true, time.Time{})
		if err != nil {
			log.Fatalf("failed to generate admin key: %v", err)
		}

		adminKey = "admin-" + key
		// store under reserved doc ID for admin
		_, err = db.Client.Collection("api_keys").Doc(adminKeyDocID).Set(ctx, map[string]interface{}{
			"key":        adminKey,
			"is_admin":   true,
			"created_at": time.Now(),
		})
		if err != nil {
			log.Fatalf("failed to store admin key: %v", err)
		}
		log.Printf("[acmutd-api] Admin key generated: %s", adminKey)
	}

	newServer := &Server{
		db:          db,
		apiKeyCache: cache.New(apiKeyCacheTTL, 10*time.Minute),
		rateLimiter: NewRateLimiter(),
		port:        port,
		adminKey:    adminKey,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
