package uploader

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	fb "firebase.google.com/go/v4"
	"github.com/acmutd/acmutd-api/internal/firebase"
	"google.golang.org/api/option"
)

var data = []string{"coursebook", "enhanced_grades", "professors"}

type UploaderService struct {
	firestoreClient *firebase.Firestore
	cloudStorage    *firebase.CloudStorage
	firebaseConfig  string
	saveEnvironment string
	config          UploaderConfig
}

type UploaderConfig struct {
	SaveEnvironment string
	InputBase       string
	classTerms      []string
	shouldGather    bool
}

func NewUploaderService(saveEnv string, gather bool) (*UploaderService, error) {
	saveEnv = strings.ToLower(saveEnv)
	if saveEnv != "dev" && saveEnv != "prod" {
		return nil, fmt.Errorf("invalid SAVE_ENVIRONMENT: %s (must be 'dev' or 'prod')", saveEnv)
	}

	classTerms := parseClassTerms(os.Getenv("CLASS_TERMS"))
	if len(classTerms) == 0 {
		return nil, fmt.Errorf("CLASS_TERMS is required (comma-separated, e.g. 24f,25s)")
	}

	return &UploaderService{
		saveEnvironment: saveEnv,
		config: UploaderConfig{
			SaveEnvironment: saveEnv,
			InputBase:       "uploader",
			classTerms:      classTerms,
			shouldGather:    gather,
		},
	}, nil
}

func parseClassTerms(raw string) []string {
	var terms []string
	for _, t := range strings.Split(raw, ",") {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" {
			terms = append(terms, t)
		}
	}
	return terms
}

func (s *UploaderService) ensureFirebaseInitialized(envHint string) error {
	env := strings.ToLower(envHint)
	if env == "" {
		env = strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	}

	if env != "dev" && env != "prod" {
		return nil
	}

	configPath, err := firebase.ResolveConfigFilename(env)
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

func (s *UploaderService) Run() error {
	if err := s.ensureFirebaseInitialized(s.config.SaveEnvironment); err != nil {
		return fmt.Errorf("failed to initialize Firebase: %w", err)
	}

	if s.config.shouldGather {
		log.Printf("Gathering data for terms: %v", s.config.classTerms)
		if err := s.gatherFromFirebase(); err != nil {
			return fmt.Errorf("gather phase failed: %w", err)
		}
		log.Println("All data gathered successfully")
	} else {
		log.Println("Skipping gather (use --gather to download from Cloud Storage)")
	}

	// TODO: implement Firestore structuring logic
	return nil
}

// gatherFromFirebase downloads term-specific coursebook and enhanced_grades files,
// plus all professor JSON from Cloud Storage into the local output directory.
func (s *UploaderService) gatherFromFirebase() error {
	if err := s.ensureUploaderDirectories(); err != nil {
		return err
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	errorCh := make(chan error, len(data))

	for _, folder := range data {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()

			destDir := filepath.Join(s.config.InputBase, dir)
			if err := s.downloadFromFolder(ctx, dir, destDir); err != nil {
				errorCh <- fmt.Errorf("failed to download from folder %s: %w", dir, err)
				return
			}

			log.Printf("Successfully downloaded data for %s", dir)
		}(folder)
	}

	wg.Wait()
	close(errorCh)

	if err := s.collectErrors(errorCh); err != nil {
		return err
	}

	log.Println("All Firebase data downloaded successfully")
	return nil
}

func (s *UploaderService) collectErrors(errorCh <-chan error) error {
	for err := range errorCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UploaderService) ensureUploaderDirectories() error {
	for _, folder := range data {
		dir := filepath.Join(s.config.InputBase, folder)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (s *UploaderService) downloadFromFolder(ctx context.Context, folderPath, outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory not found: %s", outputDir)
	}

	if err := s.ensureFirebaseInitialized(s.config.SaveEnvironment); err != nil {
		return fmt.Errorf("failed to initialize Firebase: %w", err)
	}
	fileCount, err := s.cloudStorage.DownloadFromFolder(ctx, folderPath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to download from folder %s: %w", folderPath, err)
	}

	log.Printf("Successfully downloaded %d files from %s to %s", fileCount, folderPath, outputDir)
	return nil
}
