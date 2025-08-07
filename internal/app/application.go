package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/robfig/cron"

	"Cryptoproject/config"
	"Cryptoproject/internal/adapters/client"
	"Cryptoproject/internal/adapters/storage"
	"Cryptoproject/internal/cases"
	myhttp "Cryptoproject/internal/ports/http"
)

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() {
	filePath := os.Getenv("CONFIG_FILE_PATH")
	if filePath == "" {
		filePath = "/app/config/cfg.yaml"
	}

	cfg, err := config.LoadCfg()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	connStr := "postgres://user:pass@db:5432/coinsdatabase?sslmode=disable"
	servPort := cfg.SrvPort
	client, err := client.NewClient(client.WithCustomCostIn("USD"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	storage, err := storage.NewStorage(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create storage: %v\n", err)
		os.Exit(1)
	}

	service, err := cases.NewService(storage, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create service: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		c := cron.New()
		err := c.AddFunc("@every 5m", func() { _ = service.UpdateRates(ctx) })
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cron job failed: %v\n", err)
			os.Exit(1)
		}
		c.Start()
	}()
	server, err := myhttp.NewServer(servPort, service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}
	srv := &http.Server{
		Addr:    servPort,
		Handler: server.Router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		fmt.Println("Shutting down server...")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		}
	}()

	fmt.Printf("Server running on port %s\n", servPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	wg.Wait()
}
