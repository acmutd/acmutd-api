package main

import (
	"flag"
	"log"
	"os"

	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/acmutd/acmutd-api/internal/uploader"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("[acmutd-uploader] note: could not load .env file (%v); continuing with system environment", err)
	}
	log.SetPrefix("[acmutd-uploader] ")
}

func main() {
	gather := flag.Bool("gather", false, "Download data from Cloud Storage before uploading to Firestore")
	flag.Parse()

	saveEnv := os.Getenv("SAVE_ENVIRONMENT")
	if saveEnv != "dev" && saveEnv != "prod" {
		log.Fatal("SAVE_ENVIRONMENT is required (options: dev, prod)")
	}

	log.Println("Save environment:", saveEnv)
	log.Println("Gather from Cloud Storage:", *gather)

	client := firebase.NewFBClient()
	if err := client.EnsureInitialized(saveEnv); err != nil {
		log.Fatalf("failed to initialize Firebase: %v", err)
	}

	handler, err := uploader.NewUploaderHandler(client, saveEnv, *gather)
	if err != nil {
		log.Fatalf("failed to configure upload handler: %v", err)
	}

	if err := handler.Start(); err != nil {
		log.Fatalf("uploader failed: %v", err)
	}

	log.Println("Upload to Firestore completed successfully!")
}
