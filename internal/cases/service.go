package cases

import (
	"context"
	"strings"

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

func (s *Service) GetLastRates(ctx context.Context, requestedTitles []string) ([]*entities.Coin, error) {
	if err := s.validateAndFetchTitles(ctx, requestedTitles); err != nil {
		return nil, errors.Wrap(err, "failed to preprocess requested titles")
	}

	coinsForUser, err := s.storage.GetActualCoins(ctx, requestedTitles)
	if err != nil {
		return nil, errors.Wrap(entities.ErrInternal, "failed to get coin rates")
	}

	result := make([]*entities.Coin, len(coinsForUser))
	for i, coin := range coinsForUser {
		result[i] = &coin
	}

	return result, nil
}

func (s *Service) GetAggregateRates(ctx context.Context, requestedTitles []string, aggType string) ([]*entities.Coin, error) {
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

	result := make([]*entities.Coin, len(coinsForUser))
	for i, coin := range coinsForUser {
		result[i] = &coin
	}

	return result, nil
}
func (svc *Service) UpdateRates(ctx context.Context) error {
	titles, err := svc.storage.GetCoinsList(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get coins list from storage")
	}

	currentRates, err := svc.provider.GetActualRates(ctx, titles)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve current rates from provider")
	}

	if err := svc.storage.Store(ctx, currentRates); err != nil {
		return errors.Wrap(err, "failed to store updated rates in storage")
	}

	return nil
}

func (s *Service) validateAndFetchTitles(ctx context.Context, requestedTitles []string) error {
	if len(requestedTitles) == 0 {
		return errors.Wrap(entities.ErrInvalidParam, "titles list cannot be empty")
	}

	uniqueRequestedTitles := make(map[string]bool)
	for _, title := range requestedTitles {
		uniqueRequestedTitles[title] = true
	}

	allUniqueTitles := make([]string, 0, len(uniqueRequestedTitles))
	for title := range uniqueRequestedTitles {
		allUniqueTitles = append(allUniqueTitles, title)
	}

	allNewCoins, err := s.provider.GetActualRates(ctx, allUniqueTitles)
	if err != nil {
		if strings.Contains(err.Error(), "API returned an error") {
			return errors.Wrap(entities.ErrInvalidParam, err.Error())
		}
		return errors.Wrap(err, "failed to retrieve actual rates from provider")
	}

	foundSymbols := make(map[string]struct{})
	for _, coin := range allNewCoins {
		foundSymbols[coin.Title] = struct{}{}
	}

	for _, title := range requestedTitles {
		if _, exists := foundSymbols[title]; !exists {
			return errors.Wrapf(entities.ErrInvalidParam, "coin %q does not exist or was not found in the provider", title)
		}
	}

	if err := s.storage.Store(ctx, allNewCoins); err != nil {
		return errors.Wrap(entities.ErrInternal, "failed to store new rates")
	}

	return nil
}
