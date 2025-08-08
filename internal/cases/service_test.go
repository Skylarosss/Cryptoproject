package cases_test

import (
	"context"
	"testing"

	"Cryptoproject/internal/cases"
	mocks "Cryptoproject/internal/cases/testdata"
	"Cryptoproject/internal/entities"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func setupService(t *testing.T) (*cases.Service, *mocks.MockStorage, *mocks.MockCryptoProvider) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockStorage := mocks.NewMockStorage(ctrl)
	mockProvider := mocks.NewMockCryptoProvider(ctrl)

	service, err := cases.NewService(mockStorage, mockProvider)
	require.NoError(t, err)

	return service, mockStorage, mockProvider
}

func TestService_GetLastRates_Success(t *testing.T) {
	t.Parallel()

	service, mockStorage, mockProvider := setupService(t)

	requestedTitles := []string{"BTC", "ETH"}

	mockProvider.EXPECT().
		GetActualRates(gomock.Any(), requestedTitles).
		Return([]entities.Coin{
			{Title: "BTC", Cost: 50000},
			{Title: "ETH", Cost: 3000},
		}, nil)

	mockStorage.EXPECT().
		GetActualCoins(gomock.Any(), requestedTitles).
		Return([]entities.Coin{
			{Title: "BTC", Cost: 50000},
			{Title: "ETH", Cost: 3000},
		}, nil)

	mockStorage.EXPECT().
		Store(gomock.Any(), []entities.Coin{
			{Title: "BTC", Cost: 50000},
			{Title: "ETH", Cost: 3000},
		}).Return(nil)

	rates, err := service.GetLastRates(context.Background(), requestedTitles)

	require.NoError(t, err)
	require.Len(t, rates, 2)
	require.Equal(t, float64(50000), rates[0].Cost)
	require.Equal(t, float64(3000), rates[1].Cost)
}
func TestService_GetLastRates_GetCoinListError(t *testing.T) {
	t.Parallel()

	service, mockStorage, mockProvider := setupService(t)

	requestedTitles := []string{"BTC"}

	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return(nil, entities.ErrInternal)

	mockProvider.EXPECT().
		GetActualRates(gomock.Any(), gomock.Any()).
		Return(nil, nil)

	rates, err := service.GetLastRates(context.Background(), requestedTitles)

	require.Nil(t, rates)
	require.ErrorContains(t, err, "failed to get existing coins list")
	require.ErrorIs(t, err, entities.ErrInternal)
}

func TestService_GetLastRates_GetActualRatesError(t *testing.T) {
	t.Parallel()

	service, mockStorage, mockProvider := setupService(t)

	requestedTitles := []string{"BTC", "ETH"}

	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return([]string{"BTC"}, nil)

	mockProvider.EXPECT().
		GetActualRates(gomock.Any(), []string{"ETH"}).
		Return(nil, entities.ErrInternal)

	rates, err := service.GetLastRates(context.Background(), requestedTitles)

	require.Nil(t, rates)
	require.ErrorIs(t, err, entities.ErrInternal)
	require.ErrorContains(t, err, "failed to get missing rates")
}

func TestService_GetLastRates_StoreError(t *testing.T) {
	t.Parallel()

	service, mockStorage, mockProvider := setupService(t)

	existingTitles := []string{"BTC"}

	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return(existingTitles, nil)

	mockProvider.EXPECT().
		GetActualRates(gomock.Any(), []string{"ETH"}).
		Return([]entities.Coin{{Title: "ETH", Cost: 3000}}, nil)

	mockStorage.EXPECT().
		Store(gomock.Any(), []entities.Coin{{Title: "ETH", Cost: 3000}}).
		Return(entities.ErrInternal)

	rates, err := service.GetLastRates(context.Background(), []string{"BTC", "ETH"})

	require.Nil(t, rates)
	require.ErrorIs(t, err, entities.ErrInternal)
	require.ErrorContains(t, err, "failed to store new rates")

}
func TestService_GetLastRates_GetActualCoinsError(t *testing.T) {
	t.Parallel()

	service, mockStorage, mockProvider := setupService(t)

	requestedTitles := []string{"BTC", "ETH"}
	existingTitles := []string{"BTC"}

	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return(existingTitles, nil)

	mockProvider.EXPECT().
		GetActualRates(gomock.Any(), []string{"ETH"}).
		Return([]entities.Coin{{Title: "ETH", Cost: 3000}}, nil)

	mockStorage.EXPECT().
		Store(gomock.Any(), []entities.Coin{{Title: "ETH", Cost: 3000}}).
		Return(nil)

	mockStorage.EXPECT().
		GetActualCoins(gomock.Any(), requestedTitles).
		Return(nil, entities.ErrInternal)

	rates, err := service.GetLastRates(context.Background(), []string{"BTC", "ETH"})

	require.Nil(t, rates)
	require.ErrorIs(t, err, entities.ErrInternal)
	require.ErrorContains(t, err, "failed to get coin rates")

}

func TestService_GetAggregateRates_Success(t *testing.T) {
	t.Parallel()

	service, mockStorage, mockProvider := setupService(t)

	requestedTitles := []string{"BTC", "ETH"}
	existingTitles := []string{"BTC"}
	aggType := "MAX"

	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return(existingTitles, nil)

	mockProvider.EXPECT().
		GetActualRates(gomock.Any(), []string{"ETH"}).
		Return([]entities.Coin{{Title: "ETH", Cost: 3000}}, nil)

	mockStorage.EXPECT().
		Store(gomock.Any(), []entities.Coin{{Title: "ETH", Cost: 3000}}).
		Return(nil)

	mockStorage.EXPECT().
		GetAggregateCoins(gomock.Any(), requestedTitles, aggType).
		Return([]entities.Coin{
			{Title: "BTC", Cost: 50000},
			{Title: "ETH", Cost: 3000},
		}, nil)

	rates, err := service.GetAggregateRates(context.Background(), requestedTitles, aggType)

	require.NoError(t, err)
	require.Len(t, rates, 2)
	require.Equal(t, float64(50000), rates[0].Cost)
	require.Equal(t, float64(3000), rates[1].Cost)
}

func TestNewService_StorageNilError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockCryptoProvider(ctrl)

	srv, err := cases.NewService(nil, mockProvider)

	require.Nil(t, srv)
	require.True(t, errors.Is(err, entities.ErrInvalidParam))
	require.ErrorContains(t, err, "storage not set")
}

func TestNewService_ProviderNilError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	srv, err := cases.NewService(mockStorage, nil)

	require.Nil(t, srv)
	require.True(t, errors.Is(err, entities.ErrInvalidParam))
	require.ErrorContains(t, err, "cryptoProvider not set")
}

func TestService_GetLastRates_EmptyTitles(t *testing.T) {
	t.Parallel()

	service, _, _ := setupService(t)

	rates, err := service.GetLastRates(context.Background(), []string{})

	require.Nil(t, rates)
	require.ErrorContains(t, err, "failed to preprocess requested titles")
	require.ErrorContains(t, err, "titles list cannot be empty")
	require.ErrorIs(t, err, entities.ErrInvalidParam)
}
func TestService_GetAggregateRates_EmptyTitles(t *testing.T) {
	service, _, _ := setupService(t)

	rates, err := service.GetAggregateRates(context.Background(), []string{}, "MAX")

	require.Nil(t, rates)
	require.ErrorContains(t, err, "failed to preprocess requested titles")
	require.ErrorContains(t, err, "titles list cannot be empty")
	require.ErrorIs(t, err, entities.ErrInvalidParam)
}

func TestService_GetAggregateRates_EmptyAggType(t *testing.T) {
	t.Parallel()

	service, mockStorage, _ := setupService(t)

	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return([]string{"BTC", "ETH"}, nil)

	rates, err := service.GetAggregateRates(context.Background(), []string{"BTC"}, "")

	require.Nil(t, rates)
	require.ErrorContains(t, err, "aggregation type cannot be empty")
	require.ErrorIs(t, err, entities.ErrInvalidParam)
}
func TestService_GetAggregateRates_GetAggregateCoinsError(t *testing.T) {
	t.Parallel()

	service, mockStorage, _ := setupService(t)

	requestedTitles := []string{"BTC"}
	aggType := "MAX"

	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return([]string{"BTC", "ETH"}, nil)

	mockStorage.EXPECT().
		GetAggregateCoins(gomock.Any(), requestedTitles, aggType).
		Return(nil, entities.ErrInternal)

	rates, err := service.GetAggregateRates(context.Background(), requestedTitles, aggType)

	require.Nil(t, rates)
	require.ErrorContains(t, err, "failed to get aggregate coin rates")
	require.ErrorIs(t, err, entities.ErrInternal)
}
