package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Cryptoproject/internal/adapters/client"
	"Cryptoproject/internal/adapters/storage"
	"Cryptoproject/internal/cases"
	myhttp "Cryptoproject/internal/ports/http"
)

var config struct {
	DBConnString string
	ServerPort   string
	LogLevel     string
}

func init() {
	flag.StringVar(&config.DBConnString, "dbconn", "", "Database connection string")
	flag.StringVar(&config.ServerPort, "server-port", ":8080", "Server port")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level (debug|info|warn|error)")
	flag.Parse()

	if config.DBConnString == "" {
		log.Fatal("DB connection string required!")
	}
}

func main() {
	client, err := client.NewClient(client.WithCustomCostIn("USD"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	storage, err := storage.NewStorage(config.DBConnString)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	service, err := cases.NewService(storage, client)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	srv, err := myhttp.NewServer(config.ServerPort, service)
	if err != nil {
		log.Fatalf("Failed to initialize HTTP server: %v", err)
	}

	log.Printf("Starting HTTP server on address %s...", config.ServerPort)

	srv.HttpServer.ReadTimeout = 10 * time.Second
	srv.HttpServer.WriteTimeout = 10 * time.Second

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		log.Println("Received interrupt signal, shutting down gracefully...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.HttpServer.Shutdown(ctx); err != nil {
			log.Fatalf("Server Shutdown Error: %v", err)
		}

		log.Println("Server shutdown complete.")
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed with error: %v", err)
	}

	<-done
}
