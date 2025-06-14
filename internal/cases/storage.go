package cases

import (
	"Cryptoproject/internal/entities"
	"context"
)

type Storage interface {
	Store(ctx context.Context, coins []entities.Coin) error                                          // сохраняет в хранилище
	GetCoinsList(ctx context.Context) ([]string, error)                                              // список существующих в хранилище монет
	GetActualCoins(ctx context.Context, titles []string) ([]entities.Coin, error)                    // получаем список запрашиваемых монет
	GetAggregateCoins(ctx context.Context, titles []string, aggType string) ([]entities.Coin, error) //получаем список
}
