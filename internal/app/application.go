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
	filePath := "/Users/iGamez/Desktop/Cryptoproject-1/config/cfg.yaml"

	cfg, err := config.LoadCfg(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	user := cfg.Cfg.PgUser
	pswd := cfg.Cfg.PgPswd
	host := cfg.Cfg.PgHost
	port := cfg.Cfg.PgPort
	servport := cfg.Cfg.SrvPort
	db := cfg.Cfg.PgDB
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pswd, host, port, db)

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
	server, err := myhttp.NewServer(servport, service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}
	srv := &http.Server{
		Addr:    servport,
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

	fmt.Printf("Server running on port %s\n", cfg.Cfg.SrvPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	wg.Wait()
}
