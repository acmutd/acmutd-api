package uploader

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/acmutd/acmutd-api/internal/firebase"
)

var folders = []string{"coursebook", "enhanced_grades", "professors"}

type UploaderHandler struct {
	client *firebase.FBClient
	config UploaderConfig
}

type UploaderConfig struct {
	SaveEnvironment string
	InputBase       string
	ClassTerms      []string
	ShouldGather    bool
}

func NewUploaderHandler(client *firebase.FBClient, saveEnv string, gather bool) (*UploaderHandler, error) {
	saveEnv = strings.ToLower(saveEnv)

	classTerms := parseClassTerms(os.Getenv("CLASS_TERMS"))
	if len(classTerms) == 0 {
		return nil, fmt.Errorf("CLASS_TERMS is required (comma-separated, e.g. 24f,25s)")
	}

	handler := &UploaderHandler{
		client: client,
		config: UploaderConfig{
			SaveEnvironment: saveEnv,
			ShouldGather:    gather,
			InputBase:       filepath.Join("uploader"),
			ClassTerms:      classTerms,
		},
	}
	return handler, nil
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

func (h *UploaderHandler) Start() error {
	if h.config.ShouldGather {
		log.Printf("Gathering data for terms: %v", h.config.ClassTerms)
		if err := h.client.CloudStorage().DownloadFolders(context.Background(), h.config.InputBase, folders); err != nil {
			return fmt.Errorf("gather phase failed: %w", err)
		}
		log.Println("All data gathered successfully")
	} else {
		log.Println("Skipping gather (use --gather to download from Cloud Storage)")
	}

	sections, courses, err := loadAndCombine(h.config.InputBase, h.config.ClassTerms)
	if err != nil {
		return fmt.Errorf("load and combine failed: %w", err)
	}

	log.Printf("Combined %d section documents and %d course groups across terms %v", len(sections), len(courses), h.config.ClassTerms)

	// TODO: write sections to Firestore
	return nil
}
