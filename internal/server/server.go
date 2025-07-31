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

	newServer := &Server{
		db:          db,
		apiKeyCache: cache.New(apiKeyCacheTTL, 10*time.Minute),
		rateLimiter: NewRateLimiter(),
		port:        port,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
