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

func (s *Service) UpdateCoinData(ctx context.Context) error {
	updateCoinsList, err := s.storage.GetCoinsList(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get actual coins list")
	}
	if len(updateCoinsList) == 0 {
		return nil
	}
	listCoin, err := s.provider.GetActualRates(ctx, updateCoinsList)
	if err != nil {
		return errors.Wrap(err, "failed to get actual rates from provider")
	}
	if err = s.storage.Store(ctx, listCoin); err != nil {
		return errors.Wrap(err, "failed to store actual rates")
	}
	return nil
}

func (s *Service) validateAndFetchTitles(ctx context.Context, requestedTitles []string) error {
	if len(requestedTitles) == 0 {
		return errors.Wrap(entities.ErrInvalidParam, "titles list cannot be empty")
	}

	existingTitles, err := s.storage.GetCoinsList(ctx)
	if err != nil {
		return errors.Wrap(entities.ErrInternal, "failed to get existing coins list")
	}

	missingTitles := s.findMissingTitles(requestedTitles, existingTitles)

	if len(missingTitles) > 0 {
		newCoins, err := s.provider.GetActualRates(ctx, missingTitles)
		if err != nil {
			return errors.Wrap(entities.ErrInvalidParam, "сoin does not exist or was not found in the provider")
		}

		if err := s.storage.Store(ctx, newCoins); err != nil {
			return errors.Wrap(entities.ErrInternal, "failed to store new rates")
		}
	}
	return nil
}

func (s *Service) findMissingTitles(requested, existing []string) []string {
	existingSet := make(map[string]struct{}, len(existing))
	for _, title := range existing {
		existingSet[title] = struct{}{}
	}
	missing := make([]string, 0)
	for _, title := range requested {
		if _, ok := existingSet[title]; !ok {
			missing = append(missing, title)
		}
	}
	return missing
}
