package main

import (
	"log"
	"os"

	"github.com/acmutd/acmutd-api/internal/scraper"
	"github.com/joho/godotenv"
)

func init() {
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("error loading .env file: %v\n", err)
		}
	} else {
		log.Println("Running in Docker container, skipping .env file loading")
	}
	log.SetPrefix("[acmutd-scraper] ")
}

func main() {
	outputDir := os.Getenv("OUTPUT_DIR")
	scraper := scraper.NewScraperService(outputDir)
	if err := scraper.CheckAndRunScraper(); err != nil {
		log.Fatalf("error running scraper: %v\n", err)
	}
}
