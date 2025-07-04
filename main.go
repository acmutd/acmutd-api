package main

import (
	"context"
	"log"
	"time"

	"github.com/acmutd/acmutd-api/db"
)

func main() {
	demo()
}

func demo() {
	db, err := db.NewClient(context.Background())
	if err != nil {
		log.Fatalf("error initializing firestore: %v\n", err)
	}

	// Create a dummy document
	doc := db.Client.Collection("something").Doc("test")
	doc.Set(context.Background(), map[string]interface{}{
		"name": "test",
	})

	// Read the document
	doc.Get(context.Background())

	// Sleep for 10 seconds
	time.Sleep(10 * time.Second)

	// Delete the document
	doc.Delete(context.Background())
}
