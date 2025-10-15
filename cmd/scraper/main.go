package main

import (
	"log"
	"os"

	"github.com/acmutd/acmutd-api/internal/scraper"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("[acmutd-scraper] note: could not load .env file (%v); continuing with system environment", err)
	}
	log.SetPrefix("[acmutd-scraper] ")
}

func main() {
	scraperToRun := os.Getenv("SCRAPER")
	if scraperToRun == "" {
		log.Fatal("SCRAPER environment variable is required (options: coursebook, grades, rmp-profiles, integration)")
	}

	log.Println("Running scraper:", scraperToRun)
	log.Println("Save environment:", os.Getenv("SAVE_ENVIRONMENT"))
	log.Println("")

	// Initialize scraper service with Firebase connections when required
	service, err := scraper.NewScraperService(scraperToRun)
	if err != nil {
		log.Fatalf("failed to initialize scraper service: %v", err)
	}

	var runErr error
	switch scraperToRun {
	case "coursebook", "grades", "rmp-profiles":
		// Run individual scraper (coursebook, grades, or rmp-profiles)
		runErr = service.CheckAndRunScraper()
	case "integration":
		// Run integration scraper (combines data from multiple sources)
		integrationHandler, handlerErr := scraper.NewIntegrationHandler(service)
		if handlerErr != nil {
			log.Fatalf("failed to configure integration handler: %v", handlerErr)
		}
		runErr = integrationHandler.IntegrationStart()
	default:
		log.Fatalf("Invalid SCRAPER value: %s (options: coursebook, grades, rmp-profiles, integration)", scraperToRun)
	}

	if runErr != nil {
		log.Fatalf("Scraper failed: %v\n", runErr)
	}

	log.Println("Scraper completed successfully!")
}
