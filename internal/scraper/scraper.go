package scraper

import (
	"context"
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
	scraper         string // coursebook, professor, grades
}

func NewScraperService(scraper string) *ScraperService {
	sa := option.WithCredentialsFile(os.Getenv("FIREBASE_CONFIG"))
	app, err := fb.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	firestoreClient, err := firebase.NewFirestore(context.Background(), app)
	if err != nil {
		log.Fatalf("error initializing firestore: %v\n", err)
	}

	cloudStorage, err := firebase.NewCloudStorage(context.Background(), app)
	if err != nil {
		log.Fatalf("error initializing cloud storage: %v\n", err)
	}

	return &ScraperService{
		firestoreClient: firestoreClient,
		cloudStorage:    cloudStorage,
		scraper:         scraper,
	}
}

func (s *ScraperService) CheckAndRunScraper() error {
	// Always clear output before running scraper
	if err := s.CleanupOutput(); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}

	// run the specified scraper
	runner := NewPythonRunner(s.scraper)
	err := runner.Run()
	if err != nil {
		return err
	}

	// determine whether to save locally or upload to Firebase
	// don't clear output if saving locally
	saveEnv := strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	if saveEnv != "prod" && saveEnv != "dev" {
		log.Printf("SAVE_ENVIRONMENT=%s: Data dumped locally to /out, skipping Firebase upload.", saveEnv)
		return nil
	}

	switch saveEnv {
	case "prod":
		log.Println("SAVE_ENVIRONMENT=prod: Data will be uploaded to Firebase.")
	case "dev":
		log.Println("SAVE_ENVIRONMENT=dev: Data will be uploaded to Firebase (development environment).")
	}

	// if SAVE_ENVIRONMENT is not local, upload to Firebase
	// can add dev/prod environments later
	var uploadErr error
	switch s.scraper {
	case "coursebook":
		handler := NewCoursebookHandler(s)
		uploadErr = handler.Upload(s.scraper)
	case "grades":
		handler := NewGradesHandler(s)
		uploadErr = handler.Upload(s.scraper)
	case "rmp-profiles":
		handler := NewRMPProfilesHandler(s)
		uploadErr = handler.Upload(s.scraper)
	default:
		uploadErr = fmt.Errorf("unsupported scraper type: %s", s.scraper)
	}

	// data is always saved to /scripts/{scraper}/out regardless, but we clear it if we are uploading to Firebase
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
	outputDir := "./scripts/" + s.scraper + "/out"

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
