package uploader

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/acmutd/acmutd-api/internal/firebase"
)

var folders = []string{"coursebook", "enhanced_grades", "professors"}

type UploaderService struct {
	Client *firebase.FBClient
	config UploaderConfig
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
		Client: firebase.NewFBClient(),
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

func (s *UploaderService) Run() error {
	if err := s.Client.EnsureInitialized(s.config.SaveEnvironment); err != nil {
		return fmt.Errorf("failed to initialize Firebase: %w", err)
	}

	if s.config.shouldGather {
		log.Printf("Gathering data for terms: %v", s.config.classTerms)
		if err := s.Client.CloudStorage().DownloadFolders(context.Background(), s.config.InputBase, folders); err != nil {
			return fmt.Errorf("gather phase failed: %w", err)
		}
		log.Println("All data gathered successfully")
	} else {
		log.Println("Skipping gather (use --gather to download from Cloud Storage)")
	}

	// TODO: implement Firestore structuring logic
	return nil
}
