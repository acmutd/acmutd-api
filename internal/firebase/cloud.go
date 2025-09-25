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

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/storage"
	"github.com/google/uuid"
)

type CloudStorage struct {
	*storage.Client
}

const (
	defaultBucketName = "acmutd-api.firebasestorage.app"
)

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

	bucketName := s.getBucketName()
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

func (s *CloudStorage) getBucketName() string {
	if bucketName := os.Getenv("FIREBASE_STORAGE_BUCKET"); bucketName != "" {
		return bucketName
	}
	return defaultBucketName
}
