package scraper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	fb "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/internal/firebase"
	"google.golang.org/api/option"
)

type ScraperService struct {
	firestoreClient *firebase.Firestore
	cloudStorage    *firebase.CloudStorage
	firebaseConfig  string
	scraper         string // coursebook, professor, grades, integration
}

type ServiceOption func(*ScraperService)

func NewScraperService(scraper string, opts ...ServiceOption) (*ScraperService, error) {
	if scraper == "" {
		return nil, errors.New("scraper type is required")
	}

	service := &ScraperService{scraper: scraper}

	for _, opt := range opts {
		opt(service)
	}

	return service, nil
}

func (s *ScraperService) ensureFirebaseInitialized(envHint string) error {
	env := strings.ToLower(envHint)
	if env == "" {
		env = strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	}

	if env != "dev" && env != "prod" {
		// treat anything else as local but still allow explicit FB_CONFIG
		env = "local"
	}

	configPath, err := s.resolveConfigFilename(env)
	if err != nil {
		return err
	}

	if s.firebaseConfig == configPath && s.firestoreClient != nil && s.cloudStorage != nil {
		return nil
	}

	ctx := context.Background()
	app, err := fb.NewApp(ctx, nil, option.WithCredentialsFile(configPath))
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %w", err)
	}

	firestoreClient, err := firebase.NewFirestore(ctx, app)
	if err != nil {
		return fmt.Errorf("error initializing firestore: %w", err)
	}

	cloudStorage, err := firebase.NewCloudStorage(ctx, app)
	if err != nil {
		return fmt.Errorf("error initializing cloud storage: %w", err)
	}

	s.firestoreClient = firestoreClient
	s.cloudStorage = cloudStorage
	s.firebaseConfig = configPath

	return nil
}

func (s *ScraperService) resolveConfigFilename(env string) (string, error) {
	baseName := strings.TrimSpace(os.Getenv("FB_CONFIG"))
	if baseName == "" {
		return "", errors.New("FB_CONFIG is required")
	}

	switch env {
	case "prod":
		return "prod." + baseName, nil
	case "dev", "local":
		return "dev." + baseName, nil
	default:
		return "", errors.New("invalid environment")
	}
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

	if err := s.ensureFirebaseInitialized(saveEnv); err != nil {
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
