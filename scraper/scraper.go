package scraper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/acmutd/acmutd-api/storage"
)

type ScraperService struct {
	storageClient *storage.Storage
}

func NewScraperService(storageClient *storage.Storage) *ScraperService {
	return &ScraperService{
		storageClient: storageClient,
	}
}

func (s *ScraperService) CheckAndRunScraper() error {
	outputDir := "./output"

	if s.isOutputEmpty(outputDir) {
		return fmt.Errorf("output directory is empty. Please run the docker container.")
	}

	log.Println("Output directory contains data. Skipping scraper run.")
	return nil
}

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

func (s *ScraperService) GetScrapedData() (map[string]any, error) {
	outputDir := "./output"

	if s.isOutputEmpty(outputDir) {
		return nil, fmt.Errorf("no scraped data available. Run CheckAndRunScraper() first")
	}

	data := make(map[string]any)

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(outputDir, entry.Name())
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Warning: failed to read file %s: %v", filePath, err)
				continue
			}

			// Extract term from filename (e.g., "classes_24f.json" -> "24f")
			term := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), "classes_"), ".json")
			data[term] = string(fileData)
		}
	}

	return data, nil
}

func (s *ScraperService) CleanupOutput() error {
	outputDir := "./output"

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(outputDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: failed to remove file %s: %v", filePath, err)
			}
		}
	}

	log.Println("Output directory cleaned")
	return nil
}
