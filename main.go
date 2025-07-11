package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	if err != nil {
		log.Fatalf("error loading .env file: %v\n", err)
	}

	sa := option.WithCredentialsFile(os.Getenv("FIREBASE_CONFIG"))
	app, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	storageClient, err := storage.NewStorage(context.Background(), app)
	if err != nil {
		log.Fatalf("error initializing storage client: %v\n", err)
	}

	// Create scraper service
	scraperService := scraper.NewScraperService(storageClient)

	// Check if output is empty and run scraper if needed
	log.Println("Checking if scraper needs to run...")
	if err := scraperService.CheckAndRunScraper(); err != nil {
		log.Printf("Error running scraper: %v", err)
		return
	}

	// Get the scraped data
	log.Println("Reading scraped data...")
	data, err := scraperService.GetScrapedData()
	if err != nil {
		log.Printf("Error reading scraped data: %v", err)
		return
	}

	// Process the data (example: store in Firestore)
	log.Println("Processing scraped data...")

	for term, jsonData := range data {
		bytes, err := json.Marshal(jsonData)
		if err != nil {
			log.Printf("Error marshalling data: %v", err)
			return
		}

		path := fmt.Sprintf("classes/%s.json", term)
		err = storageClient.UploadFile(context.Background(), path, bytes)
		if err != nil {
			log.Printf("Error uploading data: %v", err)
			return
		}
		log.Printf("Uploaded data for term: %s", term)
	}

	log.Println("Scraping and processing completed!")
}
