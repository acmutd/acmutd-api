package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type RMPProfilesHandler struct {
	service *ScraperService
}

func NewRMPProfilesHandler(service *ScraperService) *RMPProfilesHandler {
	return &RMPProfilesHandler{
		service: service,
	}
}

func (h *RMPProfilesHandler) Upload(scraper string) error {
	outputDir := "scripts/" + scraper + "/out"
	fmt.Println(outputDir)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory not found: %s", outputDir)
	}

	entries, err := os.ReadDir(outputDir)
	fmt.Println(entries)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	uploadCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			if err := h.uploadJSONFile(outputDir, entry.Name()); err != nil {
				log.Printf("Warning: failed to upload JSON file %s: %v", entry.Name(), err)
				continue
			}
			uploadCount++
		}
	}

	if uploadCount == 0 {
		return fmt.Errorf("no JSON files were successfully uploaded")
	}

	log.Printf("Successfully uploaded %d JSON files to cloud storage", uploadCount)
	return nil
}

func (h *RMPProfilesHandler) uploadJSONFile(outputDir, fileName string) error {
	filePath := filepath.Join(outputDir, fileName)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	cloudPath := fmt.Sprintf("rmp_data/%s", fileName)
	err = h.service.cloudStorage.UploadFile(context.Background(), cloudPath, fileData)
	if err != nil {
		return fmt.Errorf("failed to upload file to cloud storage: %w", err)
	}

	log.Printf("Successfully uploaded file: %s to cloud storage at path: %s", fileName, cloudPath)
	return nil
}
