package scraper

import (
	"log"
	"os"
)

// isOutputEmpty checks if the output directory is empty or doesn't exist
func (s *ScraperService) isOutputEmpty(outputDir string) bool {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return true
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		log.Printf("Error reading output directory: %v", err)
		return true
	}

	return len(entries) == 0
}
