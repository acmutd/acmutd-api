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
	fbclient *firebase.FBClient
	config   UploaderConfig
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
		fbclient: client,
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
		if err := h.fbclient.CloudStorage().DownloadFolders(context.Background(), h.config.InputBase, folders); err != nil {
			return fmt.Errorf("gather phase failed: %w", err)
		}
		log.Println("All data gathered successfully")
	} else {
		log.Println("Skipping gather (use --gather to download from Cloud Storage)")
	}

	lastIdx := len(h.config.ClassTerms) - 1

	for i, term := range h.config.ClassTerms {
		courses, err := loadAndCombine(h.config.InputBase, term)
		if err != nil {
			log.Printf("failed to load and combine data for term %s: %v", term, err)
			continue
		}

		log.Printf("Combined %d course groups for term %s", len(courses), term)

		if err := h.insertTermData(context.Background(), courses, term, i == lastIdx); err != nil {
			log.Printf("failed to insert data for term %s: %v", term, err)
		}
	}
	return nil
}
