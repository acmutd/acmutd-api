package main

import (
	"log"
	"os"

	"github.com/acmutd/acmutd-api/internal/scraper"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v\n", err)
	}
	log.SetPrefix("[acmutd-scraper] ")
}

func main() {
	scraperToRun := os.Getenv("SCRAPER")

	log.Println("Running scraper:", scraperToRun)
	log.Println("Saving environment:", os.Getenv("SAVE_ENVIRONMENT"))

	log.Println("Make sure you have the correct .env file set up before running the scraper.")
	log.Println("Make sure you have activated the correct virtual environment before running the scraper (source venv/bin/activate).")

	scraper := scraper.NewScraperService(scraperToRun)
	if scraperToRun != "integration" {
		err := scraper.CheckAndRunScraper()

		if err != nil {
			log.Fatalf("error running scraper: %v\n", err)
		}
	} else {
		log.Println("Integration scraper running, pulling data from Firebase.")
		err := scraper.IntegrationStart()
		if err != nil {
			log.Fatalf("error running integration scraper: %v\n", err)
		}
	}
}
