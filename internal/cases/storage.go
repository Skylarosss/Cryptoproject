package cases

import (
	"Cryptoproject/internal/entities"
	"context"
)

type Storage interface {
	Store(ctx context.Context, coins []entities.Coin) error
	GetCoinList(ctx context.Context) ([]string, error)
	GetActualCoins(ctx context.Context, titles []string) ([]entities.Coin, error)
}
