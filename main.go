package main

import (
	"context"
	"log"
	"os"
	"time"

	fb "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/api"
	"github.com/acmutd/acmutd-api/firebase"
	"github.com/acmutd/acmutd-api/scraper"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v\n", err)
	}
	log.SetPrefix("[acmutd-api] ")
}

func main() {
	ctx := context.Background()

	sa := option.WithCredentialsFile(os.Getenv("FIREBASE_CONFIG"))
	app, err := fb.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	// Initialize Firestore
	firestore, err := firebase.NewFirestore(ctx, app)
	if err != nil {
		log.Fatalf("error initializing firestore: %v\n", err)
	}

	// Initialize Cloud Storage
	storage, err := firebase.NewCloudStorage(ctx, app)
	if err != nil {
		log.Fatalf("error initializing storage: %v\n", err)
	}

	// Initialize scraper
	scraper := scraper.NewScraperService(storage, firestore)

	// Initialize API
	api := api.NewAPI(firestore)
	api.SetupRoutes()

	// Run scraper in background
	go func() {
		log.Println("Starting scraper check...")
		if err := scraper.CheckAndRunScraper(); err != nil {
			log.Printf("Error running scraper: %v", err)
		} else {
			log.Println("Scraper check completed successfully")
		}
	}()

	// FOR TESTING NOW
	if os.Getenv("CREATE_ADMIN_KEY") == "true" {

		key, err := firestore.GenerateAPIKey(ctx, 1000, 60, true, time.Now().Add(time.Hour*24*30))
		if err != nil {
			log.Printf("Failed to create admin key: %v", err)
		} else {
			log.Printf("Admin API Key: %s", key)
		}
	}

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting ACM API server on port %s", port)
	if err := api.Run(":" + port); err != nil {
		log.Fatalf("error starting server: %v\n", err)
	}
}
