package entities

import (
	"github.com/pkg/errors"
)

type Coin struct {
	Title string
	Cost  float64
}

func NewCoin(title string, cost float64) (*Coin, error) {
	if title == "" {
		return nil, errors.Wrap(ErrInvalidParam, "title cannot be empty")
	}
	if cost <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "cost must be greater than zero")
	}

	return &Coin{
		Title: title,
		Cost:  cost,
	}, nil
}
