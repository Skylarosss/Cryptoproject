package http

import (
	"context"

	"Cryptoproject/internal/entities"
)

type Service interface {
	GetLastRates(ctx context.Context, title []string) ([]*entities.Coin, error)
	GetAggregateRates(ctx context.Context, title []string, aggType string) ([]*entities.Coin, error)
}
