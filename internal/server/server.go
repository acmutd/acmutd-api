package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	fb "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/acmutd/acmutd-api/internal/server/handlers"
	"github.com/acmutd/acmutd-api/internal/server/middleware"
	"github.com/acmutd/acmutd-api/internal/server/ratelimit"
	"github.com/acmutd/acmutd-api/internal/server/router"
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
	rateLimiter *ratelimit.Limiter
	port        int
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

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error initializing firebase auth client: %v\n", err)
	}

	ctx := context.Background()

	var trackStats atomic.Bool
	appCfg, err := db.GetAppConfig(ctx)
	if err != nil {
		log.Printf("[acmutd-api] Warning: failed to load app config for trackStats: %v", err)
	} else if appCfg != nil {
		trackStats.Store(appCfg.TrackStats)
	}

	limiter := ratelimit.NewLimiter()
	limiter.StartCleanup(rateLimitCacheTTL)

	newServer := &Server{
		db:          db,
		apiKeyCache: cache.New(apiKeyCacheTTL, 10*time.Minute),
		rateLimiter: limiter,
		port:        port,
	}

	handler := handlers.New(newServer.db, &trackStats)
	middlewares := middleware.NewManager(newServer.db, newServer.apiKeyCache, newServer.rateLimiter, authClient, &trackStats)
	httpHandler := router.New(handler, middlewares)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
