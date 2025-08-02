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

	fb "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/acmutd/acmutd-api/internal/types"
	"google.golang.org/api/option"
)

type ScraperService struct {
	firestoreClient *firebase.Firestore
	outputDir       string
}

func NewScraperService(outputDir string) *ScraperService {
	sa := option.WithCredentialsFile(os.Getenv("FIREBASE_CONFIG"))
	app, err := fb.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	firestoreClient, err := firebase.NewFirestore(context.Background(), app)
	if err != nil {
		log.Fatalf("error initializing firestore: %v\n", err)
	}

	return &ScraperService{
		firestoreClient: firestoreClient,
		outputDir:       outputDir,
	}
}

func (s *ScraperService) CheckAndRunScraper() error {

	if !s.isOutputEmpty(s.outputDir) {
		log.Println("Output directory contains data. Skipping scraper run.")
		return nil
	}

	err := s.runPythonScraper()
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

// This function runs when we're inside the Docker container
func (s *ScraperService) runPythonScraper() error {
	// Check if the script exists
	if _, err := os.Stat("/app/scripts/main.py"); os.IsNotExist(err) {
		return fmt.Errorf("main.py not found in /app/scripts")
	}

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

	if s.isOutputEmpty(s.outputDir) {
		return nil, fmt.Errorf("no scraped data available. Run CheckAndRunScraper() first")
	}

	data := make(map[string][]types.Course)

	entries, err := os.ReadDir(s.outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(s.outputDir, entry.Name())
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
