package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var scrapers = []string{"coursebook", "grades", "rmp-profiles"}

type IntegrationHandler struct {
	service     *ScraperService
	config      IntegrationConfig
	directories IntegrationDirectories
}

type IntegrationConfig struct {
	Source          string
	ShouldRescrape  bool
	SaveEnvironment string
}

type IntegrationDirectories struct {
	InputBase      string
	OutputBase     string
	CoursebookOut  string
	GradesOut      string
	RMPProfilesOut string
}

func NewIntegrationHandler(service *ScraperService) (*IntegrationHandler, error) {
	if service == nil {
		return nil, errors.New("scraper service is required")
	}

	handler := &IntegrationHandler{
		service: service,
		config: IntegrationConfig{
			Source:          "local",
			ShouldRescrape:  false,
			SaveEnvironment: strings.ToLower(os.Getenv("SAVE_ENVIRONMENT")),
		},
		directories: IntegrationDirectories{
			InputBase:      filepath.Join("scripts", "integration", "in"),
			OutputBase:     filepath.Join("scripts", "integration", "out"),
			CoursebookOut:  filepath.Join("scripts", "coursebook", "out"),
			GradesOut:      filepath.Join("scripts", "grades", "out"),
			RMPProfilesOut: filepath.Join("scripts", "rmp-profiles", "out"),
		},
	}

	if err := handler.loadEnvConfig(); err != nil {
		return nil, err
	}

	return handler, nil
}

func (s *IntegrationHandler) loadEnvConfig() error {
	// Load source from env if not already set
	if source := os.Getenv("INTEGRATION_SOURCE"); source != "" {
		s.config.Source = strings.ToLower(source)
	}

	if rescrape := os.Getenv("INTEGRATION_RESCRAPE"); rescrape != "" {
		value, err := s.parseBoolEnv("INTEGRATION_RESCRAPE", s.config.ShouldRescrape)
		if err != nil {
			return err
		}
		s.config.ShouldRescrape = value
	}

	switch s.config.Source {
	case "local", "dev", "prod":
		// valid
	default:
		return fmt.Errorf("invalid INTEGRATION_SOURCE: %s (must be 'local', 'dev', or 'prod')", s.config.Source)
	}

	if s.config.Source == "local" {
		log.Println("Integration source: local filesystem")
	}

	return s.ensureBaseDirectories()
}

func (s *IntegrationHandler) ensureBaseDirectories() error {
	dirs := []string{
		s.directories.InputBase,
		filepath.Join(s.directories.InputBase, "coursebook"),
		filepath.Join(s.directories.InputBase, "grades"),
		filepath.Join(s.directories.InputBase, "rmp-profiles"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// IntegrationStart orchestrates the integration scraper workflow.
// Phases: 1) Optional rescrape, 2) Gather input data, 3) Run integration, 4) Optional upload
func (s *IntegrationHandler) IntegrationStart() error {
	log.Println("Starting Integration Scraper")

	// Phase 1: Optional Rescraping
	if s.config.ShouldRescrape {
		log.Println("\n[PHASE 1] Rescraping all data sources...")
		if err := s.rescrapeAll(); err != nil {
			return fmt.Errorf("rescrape phase failed: %w", err)
		}
	} else {
		log.Println("\n[PHASE 1] Skipping rescrape (INTEGRATION_RESCRAPE=false)")
	}

	// Phase 2: Gather Input Data
	log.Println("\n[PHASE 2] Gathering input data for integration...")
	if err := s.gatherInputData(); err != nil {
		return fmt.Errorf("data gathering phase failed: %w", err)
	}

	// Phase 3: Run Integration Script
	log.Println("\n[PHASE 3] Running integration Python script...")
	runner := NewPythonRunner("integration")
	if err := runner.Run(); err != nil {
		return fmt.Errorf("integration script failed: %w", err)
	}
	log.Println("Integration script completed successfully")

	// Phase 4: Optional Upload to Firebase
	log.Println("\n[PHASE 4] Handling output...")
	if err := s.handleIntegrationOutput(); err != nil {
		return fmt.Errorf("output handling failed: %w", err)
	}

	log.Println("Integration scraper completed successfully!")
	return nil
}

// parseBoolEnv parses a boolean environment variable with a default value
func (s *IntegrationHandler) parseBoolEnv(key string, defaultValue bool) (bool, error) {
	value := strings.ToLower(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}
	switch value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid %s: %s (must be 'true' or 'false')", key, value)
	}
}

// gatherInputData collects data from local files or Firebase based on INTEGRATION_SOURCE
func (s *IntegrationHandler) gatherInputData() error {
	log.Printf("Data source: %s", s.config.Source)

	switch s.config.Source {
	case "local":
		return s.gatherFromLocal()
	case "dev", "prod":
		return s.gatherFromFirebase()
	default:
		return fmt.Errorf("invalid INTEGRATION_SOURCE: %s (must be 'local', 'dev', or 'prod')", s.config.Source)
	}
}

// handleIntegrationOutput uploads integration results to Firebase or keeps them local
func (s *IntegrationHandler) handleIntegrationOutput() error {
	if s.config.SaveEnvironment != "prod" && s.config.SaveEnvironment != "dev" {
		log.Printf("SAVE_ENVIRONMENT=%s: Integration output saved locally to scripts/integration/out/", s.config.SaveEnvironment)
		return nil
	}

	if err := s.service.ensureFirebaseInitialized(s.config.SaveEnvironment); err != nil {
		return fmt.Errorf("failed to initialize Firebase clients: %w", err)
	}

	log.Printf("SAVE_ENVIRONMENT=%s: Uploading integration output to Firebase...", s.config.SaveEnvironment)
	if err := s.uploadIntegrationResults(); err != nil {
		return fmt.Errorf("failed to upload integration results: %w", err)
	}
	log.Println("Integration output uploaded successfully")
	return nil
}

// ensureIntegrationDirectories creates necessary directories for integration input
func (s *IntegrationHandler) ensureIntegrationDirectories() error {
	for _, scraper := range scrapers {
		dir := filepath.Join(s.directories.InputBase, scraper)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// gatherFromLocal copies data from individual scraper output directories to integration input
func (s *IntegrationHandler) gatherFromLocal() error {
	if err := s.ensureIntegrationDirectories(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	errorCh := make(chan error, len(scrapers))

	for _, scraperName := range scrapers {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			sourceDir := filepath.Join("scripts", name, "out")
			destDir := filepath.Join(s.directories.InputBase, name)

			log.Printf("Copying files: %s -> %s", sourceDir, destDir)

			if err := s.copyScraperFiles(sourceDir, destDir, name); err != nil {
				errorCh <- err
				return
			}

			log.Printf("✓ Successfully copied %s files", name)
		}(scraperName)
	}

	wg.Wait()
	close(errorCh)

	if err := s.collectErrors(errorCh); err != nil {
		return err
	}

	log.Println("✓ All local files gathered successfully")
	return nil
}

// collectErrors reads all errors from a channel and returns the first one found
func (s *IntegrationHandler) collectErrors(errorCh <-chan error) error {
	for err := range errorCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// copyScraperFiles copies all files from source to destination directory
func (s *IntegrationHandler) copyScraperFiles(sourceDir, destDir, scraperName string) error {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist for %s", sourceDir, scraperName)
	}

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", scraperName, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in %s for %s", sourceDir, scraperName)
	}

	fileCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		sourcePath := filepath.Join(sourceDir, file.Name())
		destPath := filepath.Join(destDir, file.Name())

		// Remove destination file if it exists
		if _, err := os.Stat(destPath); err == nil {
			if err := os.Remove(destPath); err != nil {
				return fmt.Errorf("failed to remove existing file %s: %w", destPath, err)
			}
		}

		if err := s.copyFile(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file.Name(), err)
		}

		fileCount++
	}

	log.Printf("  Copied %d file(s) from %s", fileCount, scraperName)
	return nil
}

func (s *IntegrationHandler) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// gatherFromFirebase downloads data from Firebase Cloud Storage to integration input
func (s *IntegrationHandler) gatherFromFirebase() error {
	if err := s.ensureIntegrationDirectories(); err != nil {
		return err
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	errorCh := make(chan error, len(scrapers))

	for _, scraperName := range scrapers {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			destDir := filepath.Join(s.directories.InputBase, name)
			log.Printf("Downloading %s from Firebase...", name)

			if err := s.downloadFromFolder(ctx, name, destDir); err != nil {
				errorCh <- fmt.Errorf("failed to download %s: %w", name, err)
				return
			}

			log.Printf("✓ Successfully downloaded %s", name)
		}(scraperName)
	}

	wg.Wait()
	close(errorCh)

	if err := s.collectErrors(errorCh); err != nil {
		return err
	}

	log.Println("✓ All Firebase data downloaded successfully")
	return nil
}

func (s *IntegrationHandler) downloadFromFolder(ctx context.Context, folderPath, outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory not found: %s", outputDir)
	}

	if err := s.service.ensureFirebaseInitialized(s.config.SaveEnvironment); err != nil {
		return fmt.Errorf("failed to initialize Firebase: %w", err)
	}
	fileCount, err := s.service.cloudStorage.DownloadFromFolder(ctx, folderPath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to download from folder %s: %w", folderPath, err)
	}

	log.Printf("Successfully downloaded %d files from %s to %s", fileCount, folderPath, outputDir)
	return nil
}

// rescrapeAll runs all scrapers and optionally uploads to Firebase
func (s *IntegrationHandler) rescrapeAll() error {
	saveEnv := strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	shouldUpload := saveEnv == "prod" || saveEnv == "dev"

	if shouldUpload {
		log.Printf("Will upload to Firebase after scraping (%s environment)", saveEnv)
	} else {
		log.Printf("Will scrape locally without uploading (%s environment)", saveEnv)
	}

	// Run scrapers sequentially (they may have dependencies)
	log.Println("\nRunning scrapers sequentially...")
	if err := s.runAllScrapers(); err != nil {
		return fmt.Errorf("scraping failed: %w", err)
	}

	// Upload concurrently if needed (uploads are independent)
	if shouldUpload {
		log.Println("\nUploading scraped data to Firebase...")
		if err := s.uploadAllScrapers(); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
	}

	log.Println("✓ Rescrape completed successfully")
	return nil
}

// runAllScrapers executes all scrapers in sequence (grades, rmp-profiles, coursebook)
func (s *IntegrationHandler) runAllScrapers() error {
	scraperOrder := []string{"grades", "rmp-profiles", "coursebook"}

	// Download existing data for term-based scrapers (to preserve old terms)
	if err := s.downloadExistingTermData(); err != nil {
		return err
	}

	var successes []string
	var errors []error

	for i, scraperName := range scraperOrder {
		log.Printf("[%d/%d] Running %s scraper...", i+1, len(scraperOrder), scraperName)

		runner := NewPythonRunner(scraperName)
		if err := runner.Run(); err != nil {
			err = fmt.Errorf("%s scraper failed: %w", scraperName, err)
			log.Printf("✗ %v", err)
			errors = append(errors, err)
			continue
		}

		successes = append(successes, scraperName)
		log.Printf("✓ %s completed successfully", scraperName)
	}

	// Handle partial failures
	if len(errors) > 0 {
		if len(successes) == 0 {
			return fmt.Errorf("all %d scrapers failed: %w", len(scraperOrder), errors[0])
		}

		failedScrapers := s.getFailedScrapers(scraperOrder, successes)
		log.Printf("\nWarning: %d/%d scrapers failed: %v", len(errors), len(scraperOrder), failedScrapers)
		log.Printf("Continuing with %d successful scraper(s): %v", len(successes), successes)
	}

	log.Printf("\n✓ Scraper execution: %d/%d successful", len(successes), len(scraperOrder))
	return nil
}

// downloadExistingTermData downloads existing coursebook and grades data to preserve old terms
func (s *IntegrationHandler) downloadExistingTermData() error {
	saveEnv := strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	if saveEnv != "prod" && saveEnv != "dev" {
		return nil // Local environment, no need to download
	}

	ctx := context.Background()
	log.Println("Downloading existing term data to preserve old terms...")

	// Download coursebook data
	if err := s.downloadFromFolder(ctx, "coursebook", "scripts/coursebook/out"); err != nil {
		log.Printf("Note: Could not download existing coursebook data: %v", err)
	}

	// Download grades data
	if err := s.downloadFromFolder(ctx, "grades", "scripts/grades/out"); err != nil {
		log.Printf("Note: Could not download existing grades data: %v", err)
	}

	return nil
}

func (s *IntegrationHandler) getFailedScrapers(allScrapers, successfulScrapers []string) []string {
	successMap := make(map[string]bool)
	for _, success := range successfulScrapers {
		successMap[success] = true
	}

	var failed []string
	for _, scraper := range allScrapers {
		if !successMap[scraper] {
			failed = append(failed, scraper)
		}
	}

	return failed
}

// uploadAllScrapers uploads data from all scrapers to Firebase concurrently
func (s *IntegrationHandler) uploadAllScrapers() error {
	type uploadJob struct {
		name    string
		handler func() error
	}

	jobs := []uploadJob{
		{"coursebook", func() error { return NewCoursebookHandler(s.service).Upload("coursebook") }},
		{"grades", func() error { return NewGradesHandler(s.service).Upload("grades") }},
		{"rmp-profiles", func() error { return NewRMPProfilesHandler(s.service).Upload("rmp-profiles") }},
	}

	var wg sync.WaitGroup
	errorCh := make(chan error, len(jobs))
	successCh := make(chan string, len(jobs))

	// Upload concurrently
	for _, job := range jobs {
		wg.Add(1)
		go func(j uploadJob) {
			defer wg.Done()

			log.Printf("Uploading %s...", j.name)

			if err := j.handler(); err != nil {
				errorCh <- fmt.Errorf("%s upload failed: %w", j.name, err)
				return
			}

			successCh <- j.name
			log.Printf("✓ %s uploaded successfully", j.name)
		}(job)
	}

	wg.Wait()
	close(errorCh)
	close(successCh)

	// Collect results
	var errors []error
	var successes []string

	for err := range errorCh {
		errors = append(errors, err)
	}

	for success := range successCh {
		successes = append(successes, success)
	}

	// Clean up output directories for successful uploads
	for _, scraperName := range successes {
		if err := s.cleanScraperOutput(scraperName); err != nil {
			log.Printf("Warning: failed to clean %s output: %v", scraperName, err)
		}
	}

	// Handle errors
	if len(errors) > 0 {
		log.Printf("\n✗ Upload failed for %d/%d scrapers", len(errors), len(jobs))
		for _, err := range errors {
			log.Printf("  - %v", err)
		}
		return fmt.Errorf("upload failed: %w", errors[0])
	}

	log.Printf("✓ All %d scrapers uploaded successfully", len(successes))
	return nil
}

// cleanScraperOutput removes files from a scraper's output directory after upload
func (s *IntegrationHandler) cleanScraperOutput(scraperName string) error {
	outputDir := filepath.Join("scripts", scraperName, "out")

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read %s output directory: %w", scraperName, err)
	}

	fileCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(outputDir, entry.Name())
		if err := os.Remove(filePath); err != nil {
			log.Printf("Warning: failed to remove %s: %v", entry.Name(), err)
		} else {
			fileCount++
		}
	}

	if fileCount > 0 {
		log.Printf("✓ Cleaned %d file(s) from %s output", fileCount, scraperName)
	}
	return nil
}

// uploadIntegrationResults uploads integration output (grades and professors) to Firebase
func (s *IntegrationHandler) uploadIntegrationResults() error {
	gradesDir := filepath.Join("scripts", "integration", "out", "grades")
	professorsDir := filepath.Join("scripts", "integration", "out", "professors")

	// Upload grades directory
	log.Println("Uploading enhanced grades data...")
	if err := s.uploadDirectory(gradesDir, "grades"); err != nil {
		return fmt.Errorf("failed to upload grades: %w", err)
	}
	log.Println("✓ Enhanced grades uploaded")

	// Upload professors directory
	log.Println("Uploading matched professor data...")
	if err := s.uploadDirectory(professorsDir, "professors"); err != nil {
		return fmt.Errorf("failed to upload professors: %w", err)
	}
	log.Println("✓ Professor data uploaded")

	return nil
}

func (s *IntegrationHandler) uploadDirectory(outputDir, cloudFolder string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory not found: %s", outputDir)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	uploadCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			if err := s.uploadFile(outputDir, entry.Name(), cloudFolder); err != nil {
				log.Printf("Warning: failed to upload file %s: %v", entry.Name(), err)
				continue
			}
			uploadCount++
		}
	}

	if uploadCount == 0 {
		return fmt.Errorf("no files were successfully uploaded from %s", outputDir)
	}

	log.Printf("Successfully uploaded %d files from %s to cloud storage", uploadCount, outputDir)
	return nil
}

func (s *IntegrationHandler) uploadFile(outputDir, fileName, cloudFolder string) error {
	filePath := filepath.Join(outputDir, fileName)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	cloudPath := fmt.Sprintf("%s/%s", cloudFolder, fileName)
	err = s.service.cloudStorage.UploadFile(context.Background(), cloudPath, fileData)
	if err != nil {
		return fmt.Errorf("failed to upload file to cloud storage: %w", err)
	}

	log.Printf("Successfully uploaded file: %s to cloud storage at path: %s", fileName, cloudPath)
	return nil
}
