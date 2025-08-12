package cases

import (
	"context"
	"log/slog"
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
	slog.Info("Starting retrieval of last rates", "requested_titles", requestedTitles)

	if err := s.ValidateAndFetchTitles(ctx, requestedTitles); err != nil {
		slog.Error("Validation failed while preprocessing requested titles", "requested_titles", requestedTitles, "err", err)
		return nil, errors.Wrap(err, "failed to preprocess requested titles")
	}

	coinsForUser, err := s.storage.GetActualCoins(ctx, requestedTitles)
	if err != nil {
		slog.Error("Failed to fetch actual coin rates", "requested_titles", requestedTitles, "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "failed to get coin rates")
	}

	result := make([]*entities.Coin, len(coinsForUser))
	for i, coin := range coinsForUser {
		result[i] = &coin
	}

	slog.Info("Retrieved latest coin rates successfully", "number_of_coins", len(result))
	return result, nil
}

func (s *Service) GetAggregateRates(ctx context.Context, requestedTitles []string, aggType string) ([]*entities.Coin, error) {
	slog.Info("Starting aggregation of coin rates", "requested_titles", requestedTitles, "agg_type", aggType)

	if err := s.ValidateAndFetchTitles(ctx, requestedTitles); err != nil {
		slog.Error("Validation failed while processing aggregate requested titles", "requested_titles", requestedTitles, "err", err)
		return nil, errors.Wrap(err, "failed to preprocess aggregate requested titles")
	}

	if aggType == "" {
		slog.Error("Aggregation type cannot be empty", "requested_titles", requestedTitles)
		return nil, errors.Wrap(entities.ErrInvalidParam, "aggregation type cannot be empty")
	}

	coinsForUser, err := s.storage.GetAggregateCoins(ctx, requestedTitles, aggType)
	if err != nil {
		slog.Error("Failed to fetch aggregated coin rates", "requested_titles", requestedTitles, "agg_type", aggType, "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "failed to get aggregate coin rates")
	}

	result := make([]*entities.Coin, len(coinsForUser))
	for i, coin := range coinsForUser {
		result[i] = &coin
	}

	slog.Info("Aggregated coin rates retrieved successfully", "number_of_coins", len(result), "agg_type", aggType)
	return result, nil
}

func (s *Service) UpdateRates(ctx context.Context) error {
	slog.Info("Updating coin rates started")

	titles, err := s.storage.GetCoinsList(ctx)
	if err != nil {
		slog.Error("Failed to get coins list from storage", "err", err)
		return errors.Wrap(err, "failed to get coins list from storage")
	}

	currentRates, err := s.provider.GetActualRates(ctx, titles)
	if err != nil {
		slog.Error("Failed to retrieve current rates from provider", "titles", titles, "err", err)
		return errors.Wrap(err, "failed to retrieve current rates from provider")
	}

	if err := s.storage.Store(ctx, currentRates); err != nil {
		slog.Error("Failed to store updated rates in storage", "titles", titles, "err", err)
		return errors.Wrap(err, "failed to store updated rates in storage")
	}

	slog.Info("Coin rates update completed successfully", "number_of_rates_updated", len(currentRates))
	return nil
}

func (s *Service) ValidateAndFetchTitles(ctx context.Context, requestedTitles []string) error {
	slog.Info("Validating and fetching requested titles", "requested_titles", requestedTitles)

	if len(requestedTitles) == 0 {
		slog.Error("Empty titles list provided", "requested_titles", requestedTitles)
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
			slog.Error("Provider API returned an error", "all_unique_titles", allUniqueTitles, "err", err)
			return errors.Wrap(entities.ErrInvalidParam, err.Error())
		}
		slog.Error("Failed to retrieve actual rates from provider", "all_unique_titles", allUniqueTitles, "err", err)
		return errors.Wrap(err, "failed to retrieve actual rates from provider")
	}

	foundSymbols := make(map[string]struct{})
	for _, coin := range allNewCoins {
		foundSymbols[coin.Title] = struct{}{}
	}

	for _, title := range requestedTitles {
		if _, exists := foundSymbols[title]; !exists {
			slog.Error("Coin not found", "missing_title", title)
			return errors.Wrapf(entities.ErrNotFound, "coin %q does not exist or was not found in the provider", title)
		}
	}

	if err := s.storage.Store(ctx, allNewCoins); err != nil {
		slog.Error("Failed to store new rates", "new_rates", allUniqueTitles, "err", err)
		return errors.Wrap(entities.ErrInternal, "failed to store new rates")
	}

	slog.Info("Title validation and fetching completed successfully", "validated_titles", requestedTitles)
	return nil
}
