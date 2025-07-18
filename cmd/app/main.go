package main

import (
	"Cryptoproject/internal/adapters/client"
	"Cryptoproject/internal/adapters/storage"
	"Cryptoproject/internal/cases"
	"context"
	"fmt"
	"log"
)

func main() {
	client, err := client.NewClient(client.WithCustomCostIn("USD"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	connStr := "postgres://user:pass@localhost:5432/coinsdatabase"
	storage, err := storage.NewStorage(connStr)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	service, err := cases.NewService(storage, client)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	requestedTitles := []string{"BTC", "ETH"}

	coins, err := service.GetLastRates(ctx, requestedTitles)
	if err != nil {
		log.Fatalf("Failed to get actual rates: %v", err)
	}

	fmt.Println("Actual Rates:")
	for _, coin := range coins {
		fmt.Printf("Title: %s, Cost: %.2f\n", coin.Title, coin.Cost)
	}

}
