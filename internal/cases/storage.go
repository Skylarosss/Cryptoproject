package cases

import (
	"Cryptoproject/internal/entities"
	"context"
)

//go:generate mockgen -source=storage.go -destination=./testdata/storage.go -package=testdata
type Storage interface {
	Store(ctx context.Context, coins []entities.Coin) error
	GetCoinsList(ctx context.Context) ([]string, error)
	GetActualCoins(ctx context.Context, titles []string) ([]entities.Coin, error)
	GetAggregateCoins(ctx context.Context, titles []string, aggType string) ([]entities.Coin, error)
}
