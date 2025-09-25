package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GradesHandler handles grades-specific scraping operations
type GradesHandler struct {
	service *ScraperService
}

// NewGradesHandler creates a new grades handler
func NewGradesHandler(service *ScraperService) *GradesHandler {
	return &GradesHandler{
		service: service,
	}
}

// Upload uploads grades data to cloud storage
func (h *GradesHandler) Upload() error {
	outputDir := "./scripts/" + h.service.scraper + "/out"

	if h.service.isOutputEmpty(outputDir) {
		return fmt.Errorf("no CSV files available in output directory")
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	ctx := context.Background()
	uploadCount := 0

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			if err := h.uploadCSVFile(ctx, outputDir, entry.Name()); err != nil {
				log.Printf("Warning: failed to upload CSV file %s: %v", entry.Name(), err)
				continue
			}
			uploadCount++
		}
	}

	if uploadCount == 0 {
		return fmt.Errorf("no CSV files were successfully uploaded")
	}

	log.Printf("Successfully uploaded %d CSV files to cloud storage", uploadCount)
	return nil
}

// uploadCSVFile uploads a single CSV file to cloud storage
func (h *GradesHandler) uploadCSVFile(ctx context.Context, outputDir, fileName string) error {
	filePath := filepath.Join(outputDir, fileName)

	// Read the CSV file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	// Upload to cloud storage with path: grades/{filename}
	cloudPath := fmt.Sprintf("grades/%s", fileName)
	err = h.service.cloudStorage.UploadFile(ctx, cloudPath, fileData)
	if err != nil {
		return fmt.Errorf("failed to upload CSV file to cloud storage: %w", err)
	}

	log.Printf("Successfully uploaded CSV file: %s to cloud storage at path: %s", fileName, cloudPath)
	return nil
}
