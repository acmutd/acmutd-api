package storage

import (
	"bytes"
	"context"
	"io"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/storage"
	"github.com/google/uuid"
)

type Storage struct {
	*storage.Client
}

var (
	bucketName = "acmutd-api.firebasestorage.app"
)

func NewStorage(ctx context.Context, app *firebase.App) (*Storage, error) {
	client, err := app.Storage(ctx)
	if err != nil {
		return nil, err
	}

	return &Storage{
		Client: client,
	}, nil
}

func (s *Storage) UploadFile(ctx context.Context, path string, data []byte) error {
	bucket, err := s.Bucket(bucketName)
	if err != nil {
		return err
	}

	object := bucket.Object(path)
	writer := object.NewWriter(ctx)

	writer.ObjectAttrs.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": uuid.New().String(),
	}

	defer writer.Close()

	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}
