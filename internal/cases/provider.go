package cases

import (
	"Cryptoproject/internal/entities"
	"context"
)

type CryptoProvider interface {
	GetActualRates(ctx context.Context, titles []string) ([]entities.Coin, error) //обновление списка монет у провайдера
}
