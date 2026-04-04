package main

import (
	"flag"
	"log"
	"os"

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
	if saveEnv == "" {
		log.Fatal("SAVE_ENVIRONMENT is required (options: dev, prod)")
	}

	log.Println("Save environment:", saveEnv)
	log.Println("Class terms:", os.Getenv("CLASS_TERMS"))
	log.Println("Gather from Cloud Storage:", *gather)

	service, err := uploader.NewUploaderService(saveEnv, *gather)
	if err != nil {
		log.Fatalf("failed to initialize uploader service: %v", err)
	}

	if err := service.Run(); err != nil {
		log.Fatalf("uploader failed: %v", err)
	}

	log.Println("Upload to Firestore completed successfully!")
}
