package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acmutd/acmutd-api/internal/server"
	"github.com/joho/godotenv"
)

func init() {
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("error loading .env file: %v\n", err)
		}
	} else {
		log.Println("Running in Docker container, skipping .env file loading")
	}
	log.SetPrefix("[acmutd-api] ")
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	done <- true
}

func main() {
	server := server.NewServer()

	done := make(chan bool, 1)
	go gracefulShutdown(server, done)

	err := server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	log.Println("Graceful shutdown complete.")
}
