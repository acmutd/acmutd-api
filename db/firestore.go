package db

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type Client struct {
	App    *firebase.App
	Client *firestore.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	configPath := os.Getenv("FIREBASE_CONFIG")
	if configPath == "" {
		return nil, fmt.Errorf("FIREBASE_CONFIG environment variable is required")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Firebase config file not found at: %s", configPath)
	}

	sa := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %w", err)
	}

	return &Client{
		App:    app,
		Client: client,
	}, nil
}
