package dto

type CoinDTO struct {
	Title string  `json: "title"`
	Cost  float64 `json: "cost"`
}
