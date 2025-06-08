package cases

import (
	"Cryptoproject/internal/entities"
	"context"
)

type Provider interface {
	GetActualRates(ctx context.Context, titles []string) (entities.Coin, error)
}
