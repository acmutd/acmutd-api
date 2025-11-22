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
	if port == 0 {
		port = 8080
	}
	log.Printf("[acmutd-api] Starting server on port %d", port)

	var configPath string
	if os.Getenv("SAVE_ENVIRONMENT") == "prod" {
		configPath = "prod." + os.Getenv("FB_CONFIG")
	} else {
		configPath = "dev." + os.Getenv("FB_CONFIG")
	}

	sa := option.WithCredentialsFile(configPath)
	app, err := fb.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	db, err := firebase.NewFirestore(context.Background(), app)
	if err != nil {
		log.Fatalf("error initializing firestore: %v\n", err)
	}

	// Delete all existing admin keys and generate a new one
	ctx := context.Background()

	if err := db.DeleteAllAdminKeys(ctx); err != nil {
		log.Printf("[acmutd-api] Warning: failed to delete existing admin keys: %v", err)
	}

	// Generate a new admin key in memory and store it in Firebase
	adminKey, err := db.GenerateAdminAPIKey(ctx)
	if err != nil {
		log.Fatalf("failed to generate admin key: %v", err)
	}

	log.Printf("[acmutd-api] New admin key generated: %s", adminKey)

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
