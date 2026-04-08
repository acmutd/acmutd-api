package firebase

import (
	"context"
	"fmt"
	"os"
	"strings"

	fb "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type FBClient struct {
	firestore    *Firestore
	cloudStorage *CloudStorage
	configPath   string
}

func NewFBClient() *FBClient {
	return &FBClient{}
}

func (c *FBClient) Firestore() *Firestore       { return c.firestore }
func (c *FBClient) CloudStorage() *CloudStorage { return c.cloudStorage }

func (c *FBClient) EnsureInitialized(envHint string) error {
	env := strings.ToLower(envHint)
	if env == "" {
		env = strings.ToLower(os.Getenv("SAVE_ENVIRONMENT"))
	}

	if env != "dev" && env != "prod" {
		env = "local"
	}

	configPath, err := ResolveConfigFilename(env)
	if err != nil {
		return err
	}

	if c.configPath == configPath && c.firestore != nil && c.cloudStorage != nil {
		return nil
	}

	ctx := context.Background()
	app, err := fb.NewApp(ctx, nil, option.WithCredentialsFile(configPath))
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %w", err)
	}

	firestoreClient, err := NewFirestore(ctx, app)
	if err != nil {
		return fmt.Errorf("error initializing firestore: %w", err)
	}

	cloudStorage, err := NewCloudStorage(ctx, app)
	if err != nil {
		return fmt.Errorf("error initializing cloud storage: %w", err)
	}

	c.firestore = firestoreClient
	c.cloudStorage = cloudStorage
	c.configPath = configPath

	return nil
}
