package cases

import (
	"context"

	"github.com/pkg/errors"

	"Cryptoproject/internal/entities"
)

type Service struct {
	storage  Storage
	provider CryptoProvider
}

func NewService(storage Storage, provider CryptoProvider) (*Service, error) {
	if storage == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "storage not set")
	}

	if provider == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "cryptoProvider not set")
	}
	return &Service{
		storage:  storage,
		provider: provider,
	}, nil
}

func (s *Service) GetLastRates(ctx context.Context, requestedTitles []string) ([]entities.Coin, error) {
	if err := s.validateAndFetchTitles(ctx, requestedTitles); err != nil {
		return nil, errors.Wrap(err, "failed to preprocess requested titles")
	}

	coinsForUser, err := s.storage.GetActualCoins(ctx, requestedTitles)
	if err != nil {
		return nil, errors.Wrap(entities.ErrInternal, "failed to get coin rates")
	}

	return coinsForUser, nil
}
func (s *Service) GetAggregateRates(ctx context.Context, requestedTitles []string, aggType string) ([]entities.Coin, error) {
	if err := s.validateAndFetchTitles(ctx, requestedTitles); err != nil {
		return nil, errors.Wrap(err, "failed to preprocess aggregate requested titles")
	}

	if aggType == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "aggregation type cannot be empty")
	}

	coinsForUser, err := s.storage.GetAggregateCoins(ctx, requestedTitles, aggType)
	if err != nil {
		return nil, errors.Wrap(entities.ErrInternal, "failed to get aggregate coin rates")
	}

	return coinsForUser, nil
}

func (s *Service) validateAndFetchTitles(ctx context.Context, requestedTitles []string) error {
	if len(requestedTitles) == 0 {
		return errors.Wrap(entities.ErrInvalidParam, "titles list cannot be empty")
	}

	uniqueRequestedTitles := make(map[string]bool)
	for _, title := range requestedTitles {
		uniqueRequestedTitles[title] = true
	}

	var allNewCoins []entities.Coin

	for title := range uniqueRequestedTitles {
		actualRate, err := s.provider.GetActualRates(ctx, []string{title})
		if err != nil || len(actualRate) == 0 {
			return errors.Wrapf(err, "coin %s does not exist or was not found in the provider", title)
		}
		allNewCoins = append(allNewCoins, actualRate...)
	}

	if err := s.storage.Store(ctx, allNewCoins); err != nil {
		return errors.Wrap(entities.ErrInternal, "failed to store new rates")
	}

	return nil
}
