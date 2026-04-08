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
	UploaderType    string // "course" or "profs"
}

func NewUploaderHandler(client *firebase.FBClient, saveEnv string, gather bool) (*UploaderHandler, error) {
	saveEnv = strings.ToLower(saveEnv)

	uploaderType := strings.ToLower(strings.TrimSpace(os.Getenv("UPLOADER")))
	if uploaderType != "course" && uploaderType != "profs" {
		return nil, fmt.Errorf("UPLOADER is required (options: course, profs)")
	}

	var classTerms []string
	if uploaderType == "course" {
		classTerms = parseClassTerms(os.Getenv("CLASS_TERMS"))
		if len(classTerms) == 0 {
			return nil, fmt.Errorf("CLASS_TERMS is required for course uploader (comma-separated, e.g. 24f,25s)")
		}
	}

	handler := &UploaderHandler{
		fbclient: client,
		config: UploaderConfig{
			SaveEnvironment: saveEnv,
			ShouldGather:    gather,
			InputBase:       filepath.Join("uploader"),
			ClassTerms:      classTerms,
			UploaderType:    uploaderType,
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
		log.Printf("Gathering data...")
		if err := h.fbclient.CloudStorage().DownloadFolders(context.Background(), h.config.InputBase, folders); err != nil {
			return fmt.Errorf("gather phase failed: %w", err)
		}
		log.Println("All data gathered successfully")
	} else {
		log.Println("Skipping gather (use --gather to download from Cloud Storage)")
	}

	switch h.config.UploaderType {
	case "course":
		return h.startCourse()
	case "profs":
		return h.startProfs()
	default:
		return fmt.Errorf("unknown uploader type: %s", h.config.UploaderType)
	}
}

func (h *UploaderHandler) startCourse() error {
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

func (h *UploaderHandler) startProfs() error {
	profs, err := loadProfessors(h.config.InputBase)
	if err != nil {
		return fmt.Errorf("failed to load professor data: %w", err)
	}

	log.Printf("Loaded %d professors", len(profs))

	h.insertProfs(context.Background(), profs)
	return nil
}
