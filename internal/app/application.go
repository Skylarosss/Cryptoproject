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
	"log/slog"
)

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() {
	cfg, err := config.LoadCfg()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	connStr := cfg.ConnStr
	servPort := cfg.SrvPort

	client, err := client.NewClient(client.WithCustomCostIn("USD"))
	if err != nil {
		slog.Error("Failed to create client", "err", err)
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	storage, err := storage.NewStorage(connStr)
	if err != nil {
		slog.Error("Failed to create storage", "err", err)
		fmt.Fprintf(os.Stderr, "Failed to create storage: %v\n", err)
		os.Exit(1)
	}

	service, err := cases.NewService(storage, client)
	if err != nil {
		slog.Error("Failed to create service", "err", err)
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
			slog.Error("Cron job setup failed", "err", err)
			fmt.Fprintf(os.Stderr, "Cron job failed: %v\n", err)
			os.Exit(1)
		}
		c.Start()
	}()

	server, err := myhttp.NewServer(servPort, service)
	if err != nil {
		slog.Error("Failed to create server", "err", err)
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
		slog.Info("Received shutdown signal")
		fmt.Println("Shutting down server...")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown error", "err", err)
			fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
		}
	}()

	slog.Info("Server starting", "port", servPort)
	fmt.Printf("Server running on port %s\n", servPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("ListenAndServe error", "err", err)
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	wg.Wait()
}
