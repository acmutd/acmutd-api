package scraper

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/acmutd/acmutd-api/internal/firebase"
)

type ScraperService struct {
	fbClient *firebase.FBClient
	scraper  string // coursebook, professor, grades, integration
}

type ServiceOption func(*ScraperService)

func NewScraperService(scraper string, opts ...ServiceOption) (*ScraperService, error) {
	if scraper == "" {
		return nil, errors.New("scraper type is required")
	}

	service := &ScraperService{
		fbClient: firebase.NewFBClient(),
		scraper:  scraper,
	}

	for _, opt := range opts {
		opt(service)
	}

	return service, nil
}

func (s *ScraperService) CheckAndRunScraper() error {
	if s.scraper == "" {
		return errors.New("scraper type is required")
	}

	// Always clear output before running scraper
	if err := s.CleanupOutput(); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}

	// run the specified scraper
	runner := NewPythonRunner(s.scraper)
	if err := runner.Run(); err != nil {
		return err
	}

	// determine whether to save locally or upload to Firebase
	// don't clear output if saving locally
	saveEnv := strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	if saveEnv != "prod" && saveEnv != "dev" {
		log.Printf("SAVE_ENVIRONMENT=%s: Data dumped locally to /out, skipping Firebase upload.", saveEnv)
		return nil
	}

	if err := s.fbClient.EnsureInitialized(saveEnv); err != nil {
		return fmt.Errorf("failed to initialize cloud storage: %w\nOutput not deleted to prevent data loss", err)
	}

	switch saveEnv {
	case "prod":
		log.Println("SAVE_ENVIRONMENT=prod: Data will be uploaded to Firebase (prod environment)")
	case "dev":
		log.Println("SAVE_ENVIRONMENT=dev: Data will be uploaded to Firebase (development environment)")
	}

	var uploadErr error
	switch s.scraper {
	case "coursebook":
		uploadErr = NewCoursebookHandler(s).Upload(s.scraper)
	case "grades":
		uploadErr = NewGradesHandler(s).Upload(s.scraper)
	case "rmp-profiles":
		uploadErr = NewRMPProfilesHandler(s).Upload(s.scraper)
	default:
		uploadErr = fmt.Errorf("unsupported scraper type: %s", s.scraper)
	}

	cleanupErr := s.CleanupOutput()
	if uploadErr != nil {
		return uploadErr
	}
	if cleanupErr != nil {
		return fmt.Errorf("upload succeeded but failed to clean output directory: %w", cleanupErr)
	}
	return nil
}

func (s *ScraperService) CleanupOutput() error {
	outputDir := "scripts/" + s.scraper + "/out"

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
