package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/acmutd/acmutd-api/firebase"
	"github.com/acmutd/acmutd-api/types"
)

type ScraperService struct {
	storageClient   *firebase.CloudStorage
	firestoreClient *firebase.Firestore
}

func NewScraperService(storageClient *firebase.CloudStorage, firestoreClient *firebase.Firestore) *ScraperService {
	return &ScraperService{
		storageClient:   storageClient,
		firestoreClient: firestoreClient,
	}
}

func (s *ScraperService) CheckAndRunScraper() error {
	var outputDir string
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		outputDir = "/app/output"
	} else {
		outputDir = "./output"
	}

	if !s.isOutputEmpty(outputDir) {
		log.Println("Output directory contains data. Skipping scraper run.")
		return nil
	}

	err := s.runScraperContainer()
	if err != nil {
		return err
	}

	data, err := s.GetScrapedData()
	if err != nil {
		return err
	}

	terms := strings.Split(os.Getenv("CLASS_TERMS"), ",")
	s.firestoreClient.InsertTerms(context.Background(), terms)

	for term, courses := range data {
		s.firestoreClient.InsertClassesWithIndexes(context.Background(), courses, term)
	}

	return nil
}

func (s *ScraperService) runScraperContainer() error {
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return s.runPythonScraper()
	}

	// We're outside Docker, use docker compose to run the scraper
	cmd := exec.Command("docker", "compose", "run", "--rm", "--no-TTY", "scraper")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Starting scraper container...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run scraper container: %w", err)
	}

	// Wait a bit for files to be written
	time.Sleep(5 * time.Second)

	log.Println("Scraper container completed successfully")
	return nil
}

// This function runs when we're inside the Docker container
func (s *ScraperService) runPythonScraper() error {
	cmd := exec.Command("python", "main.py")
	cmd.Dir = "/app/scripts"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PYTHONPATH=/app/scripts")

	log.Println("Running Python scraper from directory:", cmd.Dir)
	log.Println("Python command:", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run Python scraper: %w", err)
	}

	log.Println("Python scraper completed successfully")
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

func (s *ScraperService) GetScrapedData() (map[string][]types.Course, error) {
	var outputDir string
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		outputDir = "/app/output"
	} else {
		outputDir = "./output"
	}

	if s.isOutputEmpty(outputDir) {
		return nil, fmt.Errorf("no scraped data available. Run CheckAndRunScraper() first")
	}

	data := make(map[string][]types.Course)

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
			var courses []types.Course
			if err := json.Unmarshal(fileData, &courses); err != nil {
				log.Printf("Warning: failed to unmarshal file %s: %v", filePath, err)
				continue
			}
			data[term] = courses
		}
	}

	return data, nil
}

func (s *ScraperService) CleanupOutput() error {
	outputDir := "/app/output"

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
