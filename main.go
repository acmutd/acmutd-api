package main

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/scraper"
	"github.com/acmutd/acmutd-api/storage"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	err := godotenv.Load()
	ctx := context.Background()
	if err != nil {
		log.Fatalf("error loading .env file: %v\n", err)
	}

	sa := option.WithCredentialsFile(os.Getenv("FIREBASE_CONFIG"))
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	storage, err := storage.NewStorage(ctx, app)
	if err != nil {
		log.Fatalf("error initializing storage: %v\n", err)
	}

	scraper := scraper.NewScraperService(storage)

	log.Println("Starting scraper check...")
	if err := scraper.CheckAndRunScraper(); err != nil {
		log.Printf("Error running scraper: %v", err)
	} else {
		log.Println("Scraper check completed successfully")
	}

}
