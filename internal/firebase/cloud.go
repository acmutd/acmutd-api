package firebase

import (
	"bytes"
	"context"
	"io"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/storage"
	"github.com/google/uuid"
)

type CloudStorage struct {
	*storage.Client
}

var (
	bucketName = "acmutd-api.firebasestorage.app"
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
