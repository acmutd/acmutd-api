package scraper

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	scrapers  = []string{"coursebook", "grades", "rmp-profiles"}
	outputDir = "scripts/integration/in"
)

func (s *ScraperService) IntegrationStart() error {
	integrationMode := strings.ToLower(os.Getenv("INTEGRATION_SOURCE"))
	integrationRescrape := strings.ToLower(os.Getenv("INTEGRATION_RESCRAPE"))

	shouldRescrape := false
	if integrationRescrape != "" {
		switch integrationRescrape {
		case "true":
			shouldRescrape = true
		case "false":
			shouldRescrape = false
		default:
			return fmt.Errorf("invalid INTEGRATION_RESCRAPE: %s (must be 'true' or 'false')", integrationRescrape)
		}
	}

	var err error

	if shouldRescrape {
		log.Println("INTEGRATION_RESCRAPE=true: Running scrapers before data processing...")
		err = s.RescrapeStart()
		if err != nil {
			return fmt.Errorf("rescrape failed: %w", err)
		}
	}

	// Then handle data source based on INTEGRATION_SOURCE
	switch integrationMode {
	case "local":
		err = s.LocalStart()
	case "dev":
		fallthrough
	case "prod":
		err = s.FirebaseStart()
	case "":
		log.Println("No INTEGRATION_SOURCE specified, defaulting to 'local'")
		err = s.LocalStart()
	default:
		return fmt.Errorf("invalid INTEGRATION_SOURCE: %s (must be 'local', 'dev', or 'prod')", integrationMode)
	}

	// TODO: Run the integration scraper

	return err
}

func (s *ScraperService) LocalStart() error {
	localOutBasePath := "scripts/"

	for _, scraper := range scrapers {
		dir := filepath.Join(outputDir, scraper)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Printf("Ensured directory exists: %s", dir)
	}

	var wg sync.WaitGroup
	errorCh := make(chan error, len(scrapers))

	for _, scraper := range scrapers {
		wg.Add(1)
		go func(scraperName string) {
			defer wg.Done()

			sourceDir := filepath.Join(localOutBasePath, scraperName, "out")
			destDir := filepath.Join(outputDir, scraperName)

			log.Printf("Processing local files for %s: %s -> %s", scraperName, sourceDir, destDir)

			if err := s.processLocalScraperFiles(sourceDir, destDir, scraperName); err != nil {
				errorCh <- fmt.Errorf("failed to process %s files: %w", scraperName, err)
				return
			}

			log.Printf("Successfully processed local files for %s", scraperName)
		}(scraper)
	}

	wg.Wait()
	close(errorCh)

	for err := range errorCh {
		if err != nil {
			return err
		}
	}

	log.Println("All local file processing completed successfully")
	return nil
}

func (s *ScraperService) processLocalScraperFiles(sourceDir, destDir, scraperName string) error {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		log.Printf("Warning: source directory %s does not exist, skipping %s", sourceDir, scraperName)
		return nil
	}

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", sourceDir, err)
	}

	if len(files) == 0 {
		log.Printf("Warning: no files found in %s for scraper %s", sourceDir, scraperName)
		return nil
	}

	log.Printf("Found %d files to process for %s", len(files), scraperName)

	fileCount := 0
	for _, file := range files {
		if file.IsDir() {
			log.Printf("Skipping directory: %s", file.Name())
			continue
		}

		sourcePath := filepath.Join(sourceDir, file.Name())
		destPath := filepath.Join(destDir, file.Name())

		if _, err := os.Stat(destPath); err == nil {
			log.Printf("Warning: destination file %s already exists, overwriting", destPath)
			if err := os.Remove(destPath); err != nil {
				return fmt.Errorf("failed to remove existing file %s: %w", destPath, err)
			}
		}

		if err := s.copyFile(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy file %s to %s: %w", sourcePath, destPath, err)
		}

		log.Printf("Copied: %s -> %s", file.Name(), destPath)
		fileCount++
	}

	log.Printf("Successfully processed %d files for %s", fileCount, scraperName)
	return nil
}

func (s *ScraperService) copyFile(src, dst string) error {
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

func (s *ScraperService) FirebaseStart() error {
	ctx := context.Background()
	for _, scraper := range scrapers {
		dir := filepath.Join(outputDir, scraper)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Printf("Ensured directory exists: %s", dir)
	}

	var wg sync.WaitGroup
	errorCh := make(chan error, len(scrapers))

	for _, scraper := range scrapers {
		wg.Add(1)
		go func(scraperName string) {
			defer wg.Done()
			dir := filepath.Join(outputDir, scraperName)
			log.Printf("Starting download for %s to %s", scraperName, dir)
			if err := s.downloadFromFolder(ctx, scraperName, dir); err != nil {
				errorCh <- fmt.Errorf("failed to download %s: %w", scraperName, err)
				return
			}
			log.Printf("Completed download for %s", scraperName)
		}(scraper)
	}

	wg.Wait()
	close(errorCh)

	for err := range errorCh {
		if err != nil {
			return err
		}
	}

	log.Println("All downloads completed successfully")
	return nil
}

func (s *ScraperService) downloadFromFolder(ctx context.Context, folderPath, outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory not found: %s", outputDir)
	}

	fileCount, err := s.cloudStorage.DownloadFromFolder(ctx, folderPath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to download from folder %s: %w", folderPath, err)
	}

	log.Printf("Successfully downloaded %d files from %s to %s", fileCount, folderPath, outputDir)
	return nil
}

func (s *ScraperService) RescrapeStart() error {
	log.Println("Starting rescrape operation for all scrapers...")

	saveEnv := strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	shouldUpload := saveEnv == "prod" || saveEnv == "dev"

	if shouldUpload {
		log.Printf("SAVE_ENVIRONMENT=%s: Will upload to Firebase after scraping", saveEnv)
	} else {
		log.Printf("SAVE_ENVIRONMENT=%s: Will only scrape locally, skipping upload", saveEnv)
	}

	log.Println("Running all scrapers sequentially...")
	if err := s.runAllScrapers(); err != nil {
		return fmt.Errorf("scraping phase failed: %w", err)
	}

	if shouldUpload {
		log.Println("Uploading all data concurrently...")
		if err := s.uploadAllScrapers(); err != nil {
			return fmt.Errorf("upload phase failed: %w", err)
		}
	}

	log.Println("Rescrape operation completed successfully")

	return nil
}

func (s *ScraperService) runAllScrapers() error {
	scraperOrder := []string{"grades", "rmp-profiles", "coursebook"}
	ctx := context.Background()

	saveEnv := strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	// We need to grab all existing terms for coursebook and grades (if not local)
	// Do not do this for rmp-profiles because it is not a term-based scraper
	if saveEnv == "prod" || saveEnv == "dev" {
		err := s.downloadFromFolder(ctx, "coursebook", "scripts/coursebook/out")
		if err != nil {
			return fmt.Errorf("failed to download coursebook files: %w", err)
		}

		err = s.downloadFromFolder(ctx, "grades", "scripts/grades/out")
		if err != nil {
			return fmt.Errorf("failed to download grades files: %w", err)
		}
	}

	var successes []string
	var errors []error

	for _, scraperName := range scraperOrder {
		log.Printf("Starting scraper %d/%d: %s", len(successes)+1, len(scraperOrder), scraperName)

		runner := NewPythonRunner(scraperName)
		if err := runner.Run(); err != nil {
			err = fmt.Errorf("failed to run %s scraper: %w", scraperName, err)
			log.Printf("Scraping error: %v", err)
			errors = append(errors, err)
			continue
		}

		successes = append(successes, scraperName)
		log.Printf("Successfully completed scraper %d/%d: %s", len(successes), len(scraperOrder), scraperName)

	}

	if len(errors) > 0 {
		log.Printf("Scraping completed with %d errors, %d successes", len(errors), len(successes))
		log.Printf("Failed scrapers: %v", s.getFailedScrapers(scraperOrder, successes))
		log.Printf("Successful scrapers: %v", successes)

		if len(successes) == 0 {
			return fmt.Errorf("all scrapers failed, first error: %w", errors[0])
		}

		log.Printf("Warning: %d scrapers failed, but %d succeeded. Continuing with available data.", len(errors), len(successes))
	}

	log.Printf("Scraper execution completed: %d/%d successful", len(successes), len(scraperOrder))
	return nil
}

func (s *ScraperService) getFailedScrapers(allScrapers, successfulScrapers []string) []string {
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

func (s *ScraperService) uploadAllScrapers() error {
	type uploadJob struct {
		name    string
		handler func() error
	}

	jobs := []uploadJob{
		{"coursebook", func() error { return NewCoursebookHandler(s).Upload("coursebook") }},
		{"grades", func() error { return NewGradesHandler(s).Upload("grades") }},
		{"rmp-profiles", func() error { return NewRMPProfilesHandler(s).Upload("rmp-profiles") }},
	}

	var wg sync.WaitGroup
	errorCh := make(chan error, len(jobs))
	successCh := make(chan string, len(jobs))

	for _, job := range jobs {
		wg.Add(1)
		go func(j uploadJob) {
			defer wg.Done()

			log.Printf("Starting upload for: %s", j.name)

			if err := j.handler(); err != nil {
				errorCh <- fmt.Errorf("failed to upload %s: %w", j.name, err)
				return
			}

			successCh <- j.name
			log.Printf("Successfully uploaded: %s", j.name)
		}(job)
	}

	wg.Wait()
	close(errorCh)
	close(successCh)

	var errors []error
	var successes []string

	for err := range errorCh {
		if err != nil {
			errors = append(errors, err)
		}
	}

	for success := range successCh {
		successes = append(successes, success)
	}

	for _, scraperName := range successes {
		if err := s.cleanScraperOutput(scraperName); err != nil {
			log.Printf("Warning: failed to clean output for %s: %v", scraperName, err)
		} else {
			log.Printf("Cleaned output directory for: %s", scraperName)
		}
	}

	if len(errors) > 0 {
		log.Printf("Upload completed with %d errors, %d successes", len(errors), len(successes))
		for _, err := range errors {
			log.Printf("Upload error: %v", err)
		}
		return fmt.Errorf("upload failed with %d errors, first error: %w", len(errors), errors[0])
	}

	log.Printf("All uploads completed successfully (%d scrapers)", len(successes))
	return nil
}

func (s *ScraperService) cleanScraperOutput(scraperName string) error {
	outputDir := filepath.Join("scripts", scraperName, "out")

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		log.Printf("Output directory %s does not exist for %s, nothing to clean", outputDir, scraperName)
		return nil
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory %s: %w", outputDir, err)
	}

	fileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(outputDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: failed to remove file %s: %v", filePath, err)
			} else {
				fileCount++
			}
		}
	}

	if fileCount > 0 {
		log.Printf("Cleaned %d files from output directory: %s", fileCount, outputDir)
	}
	return nil
}
