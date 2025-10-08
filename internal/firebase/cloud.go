package firebase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	goStorage "cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type CloudStorage struct {
	*storage.Client
}

func GetDefaultBucketName() string {
	saveEnvironment := os.Getenv("SAVE_ENVIRONMENT")
	if saveEnvironment == "local" || saveEnvironment == "dev" {
		return "acmutd-api-dev.firebasestorage.app"
	} else if saveEnvironment == "prod" {
		return "acmutd-api.firebasestorage.app"
	}
	return ""
}

func NewCloudStorage(ctx context.Context, app *firebase.App) (*CloudStorage, error) {
	client, err := app.Storage(ctx)
	if err != nil {
		return nil, err
	}

	return &CloudStorage{
		Client: client,
	}, nil
}

func (s *CloudStorage) UploadFile(ctx context.Context, path string, data []byte) error {
	if err := s.validateUpload(path, data); err != nil {
		return fmt.Errorf("upload validation failed: %w", err)
	}

	bucketName := GetDefaultBucketName()
	bucket, err := s.Bucket(bucketName)
	if err != nil {
		return fmt.Errorf("failed to get storage bucket '%s': %w", bucketName, err)
	}

	object := bucket.Object(path)
	writer := object.NewWriter(ctx)
	defer writer.Close()

	contentType := s.detectContentType(path)
	writer.ObjectAttrs.ContentType = contentType

	writer.ObjectAttrs.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": uuid.New().String(),
	}

	reader := bytes.NewReader(data)
	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to upload file data: %w", err)
	}

	return nil
}

func (s *CloudStorage) DownloadFromFolder(ctx context.Context, folderPath, outputDir string) (int, error) {
	bucketName := GetDefaultBucketName()
	bucket, err := s.Bucket(bucketName)
	if err != nil {
		return 0, fmt.Errorf("failed to get storage bucket '%s': %w", bucketName, err)
	}

	var objectInfos []struct {
		name string
		size int64
	}

	objects := bucket.Objects(ctx, &goStorage.Query{
		Prefix: folderPath,
	})

	for {
		object, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("failed to iterate objects: %w", err)
		}

		// Skip directories (objects ending with '/')
		if strings.HasSuffix(object.Name, "/") {
			continue
		}

		objectInfos = append(objectInfos, struct {
			name string
			size int64
		}{object.Name, object.Size})
	}

	if len(objectInfos) == 0 {
		log.Printf("No files found in folder: %s", folderPath)
		return 0, nil
	}

	log.Printf("Found %d files to download from %s", len(objectInfos), folderPath)

	// Download files concurrently with worker pool
	const maxWorkers = 5
	workChan := make(chan struct {
		name string
		size int64
	}, len(objectInfos))
	errorChan := make(chan error, len(objectInfos))
	completedChan := make(chan string, len(objectInfos))

	// Start workers
	for i := 0; i < maxWorkers && i < len(objectInfos); i++ {
		go s.downloadWorker(ctx, bucket, outputDir, folderPath, workChan, errorChan, completedChan)
	}

	// Send work to workers
	for _, obj := range objectInfos {
		workChan <- obj
	}
	close(workChan)

	// Collect results
	var errors []error
	fileCount := 0

	for i := 0; i < len(objectInfos); i++ {
		select {
		case err := <-errorChan:
			if err != nil {
				errors = append(errors, err)
			}
		case fileName := <-completedChan:
			log.Printf("Downloaded: %s", fileName)
			fileCount++
		}
	}

	if len(errors) > 0 {
		return fileCount, fmt.Errorf("failed to download %d files: %v", len(errors), errors[0])
	}

	return fileCount, nil
}

// validateUpload performs input validation for file uploads
func (s *CloudStorage) validateUpload(path string, data []byte) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if len(data) == 0 {
		return fmt.Errorf("file data cannot be empty")
	}

	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return fmt.Errorf("invalid file path: contains unsafe characters")
	}

	return nil
}

func (s *CloudStorage) detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		log.Printf("detected MIME type: %s for file: %s", mimeType, path)
		return mimeType
	}

	switch ext {
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	case ".xlsx", ".xls":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

// downloadWorker processes download jobs concurrently
func (s *CloudStorage) downloadWorker(ctx context.Context, bucket *goStorage.BucketHandle, outputDir, folderPath string, workChan <-chan struct {
	name string
	size int64
}, errorChan chan<- error, completedChan chan<- string) {
	for work := range workChan {
		if err := s.downloadSingleFile(ctx, bucket, work.name, outputDir, folderPath); err != nil {
			errorChan <- fmt.Errorf("failed to download %s: %w", work.name, err)
		} else {
			completedChan <- work.name
		}
	}
}

// downloadSingleFile downloads a single file from cloud storage
func (s *CloudStorage) downloadSingleFile(ctx context.Context, bucket *goStorage.BucketHandle, objectName, outputDir, folderPath string) error {
	objBucket := bucket.Object(objectName)
	reader, err := objBucket.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()

	objectData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read object data: %w", err)
	}

	// Extract relative path by removing folder prefix
	relativePath := strings.TrimPrefix(objectName, folderPath)
	relativePath = strings.TrimPrefix(relativePath, "/")

	// Ensure we don't have empty relative path
	if relativePath == "" {
		return fmt.Errorf("invalid object path: %s", objectName)
	}

	filePath := filepath.Join(outputDir, relativePath)

	// Create directory structure if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(filePath, objectData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
