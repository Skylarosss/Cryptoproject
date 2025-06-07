package entities_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"Cryptoproject/internal/entities"
)

func Test_NewCoin_Success(t *testing.T) {
	t.Parallel()
	validTitle := "BTC"
	validCost := 106000.0
	coin, err := entities.NewCoin(validTitle, validCost)
	require.NoError(t, err)
	require.Equal(t, &entities.Coin{
		Title: validTitle,
		Cost:  validCost,
	}, coin)
}
func Test_NewCoin_EmptyTitle(t *testing.T) {
	t.Parallel()
	invalidTitle := ""
	validCost := 106000.0
	coin, err := entities.NewCoin(invalidTitle, validCost)
	require.ErrorIs(t, err, entities.ErrInvalidParam)
	require.Nil(t, coin)
}
func Test_NewCoin_InvalidCost(t *testing.T) {
	t.Parallel()
	validTitle := "BTC"
	invalidCost := 0.0
	coin, err := entities.NewCoin(validTitle, invalidCost)
	require.ErrorIs(t, err, entities.ErrInvalidParam)
	require.Nil(t, coin)
}
