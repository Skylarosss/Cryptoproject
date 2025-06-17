package cases_test

import (
	"context"
	"testing"

	"Cryptoproject/internal/cases"
	"Cryptoproject/internal/entities"
	"Cryptoproject/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_GetLastRates_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockProvider := mocks.NewMockCryptoProvider(ctrl)

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
		Return([]entities.Coin{
			{Title: "BTC", Cost: 50000},
			{Title: "ETH", Cost: 3000},
		}, nil)

	service, err := cases.NewService(mockStorage, mockProvider)
	require.NoError(t, err)

	rates, err := service.GetLastRates(context.Background(), requestedTitles)

	assert.NoError(t, err)
	assert.Len(t, rates, 2)
	assert.Equal(t, 50000.0, rates[0].Cost)
	assert.Equal(t, 3000.0, rates[1].Cost)
}
