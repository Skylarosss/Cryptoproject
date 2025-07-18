package cases

import (
	"Cryptoproject/internal/entities"
	"context"
)

//go:generate mockgen -source=provider.go -destination=./testdata/provider.go -package=testdata
type CryptoProvider interface {
	GetActualRates(ctx context.Context, titles []string) ([]entities.Coin, error)
}
