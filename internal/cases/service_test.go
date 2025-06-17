package cases_test

import (
	"context"
	"testing"

	"Cryptoproject/internal/cases"
	"Cryptoproject/internal/entities"
	"Cryptoproject/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestService_GetLastRates_Success(t *testing.T) {
	// 1. Инициализация
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2. Создаем моки зависимостей
	mockStorage := mocks.NewMockStorage(ctrl)
	mockProvider := mocks.NewMockCryptoProvider(ctrl)

	// 3. Настраиваем ожидания для моков
	requestedTitles := []string{"BTC", "ETH"}
	existingTitles := []string{"BTC"} // ETH отсутствует в хранилище

	// Ожидаем вызов GetCoinsList
	mockStorage.EXPECT().
		GetCoinsList(gomock.Any()).
		Return(existingTitles, nil).
		Times(1)

	// Ожидаем запрос недостающих курсов (ETH)
	mockProvider.EXPECT().
		GetActualRates(gomock.Any(), []string{"ETH"}).
		Return([]entities.Coin{{Title: "ETH", Cost: 3000}}, nil).
		Times(1)

	// Ожидаем сохранение новых курсов
	mockStorage.EXPECT().
		Store(gomock.Any(), []entities.Coin{{Title: "ETH", Cost: 3000}}).
		Return(nil).
		Times(1)

	// Ожидаем запрос актуальных курсов
	mockStorage.EXPECT().
		GetActualCoins(gomock.Any(), requestedTitles).
		Return([]entities.Coin{
			{Title: "BTC", Cost: 50000},
			{Title: "ETH", Cost: 3000},
		}, nil).
		Times(1)

	// 4. Создаем тестируемый сервис
	service, err := cases.NewService(mockStorage, mockProvider)
	assert.NoError(t, err)

	// 5. Вызываем метод
	rates, err := service.GetLastRates(context.Background(), requestedTitles)

	// 6. Проверяем результаты
	assert.NoError(t, err)
	assert.Len(t, rates, 2)
	assert.Equal(t, 50000.0, rates[0].Cost)
	assert.Equal(t, 3000.0, rates[1].Cost)
}
