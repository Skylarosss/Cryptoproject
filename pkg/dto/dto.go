package dto

// ResponseDTO model contains a collection of CoinDTO objects representing the final response.
// swagger:model
type ResponseDTO struct {
	Coins []CoinDTO `json:"coins"`
}

// CoinDTO model represents detailed information about a single cryptocurrency.
// swagger:model
type CoinDTO struct {
	Title string  `json:"title"`
	Cost  float64 `json:"cost"`
}

// ErrorResponseDTO model defines the format of an error response when something goes wrong.
// swagger:model
type ErrorResponseDTO struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RequestDTO model specifies the input data needed for retrieving rates.
// swagger:model
type RequestDTO struct {
	Titles  []string `json:"titles"`
	AggType string   `json:"aggType"`
}
