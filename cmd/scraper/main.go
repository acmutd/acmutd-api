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
	scraper := scraper.NewScraperService(scraperToRun)
	if err := scraper.CheckAndRunScraper(); err != nil {
		log.Fatalf("error running scraper: %v\n", err)
	}
}
