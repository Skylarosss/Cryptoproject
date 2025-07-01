package main

import (
	"Cryptoproject/internal/adapters/client"
	"context"
	"fmt"
)

func main() {
	client, err := client.NewClient(client.WithCustomCostIn("USD"))
	if err != nil {
		panic(err)
	}
	coins, err := client.GetActualRates(context.TODO(), []string{"BTC", "ETH", "SOL", "STRK", "ZRO"})
	if err != nil {
		panic(err)
	}
	fmt.Println(coins)
}
