package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

type Client struct {
	app    *firebase.App
	Client *firestore.Client
}

func NewClient(ctx context.Context, app *firebase.App) (*Client, error) {

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore client: %w", err)
	}

	return &Client{
		app:    app,
		Client: client,
	}, nil
}
