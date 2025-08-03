package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"Cryptoproject/internal/entities"
	"Cryptoproject/pkg/dto"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type Server struct {
	HttpServer *http.Server
	Service    Service
	Router     *chi.Mux
}

func NewServer(addr string, srv Service) (*Server, error) {
	if addr == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "address cannot be empty")
	}
	if srv == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "service not provided")
	}

	router := chi.NewRouter()
	srvInstance := &Server{
		HttpServer: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		Service: srv,
		Router:  router,
	}

	// @Title Cryptocurrency Rates API
	// @Version 1.0
	// @Description This API provides cryptocurrency rate information.
	// @Host localhost:8080
	// @BasePath /
	// @schemes http
	// @produces json
	// @consumes json

	srvInstance.Router.Post("/rates/last", srvInstance.getLastRates)
	srvInstance.Router.Post("/rates/aggregate", srvInstance.getAggregateRates)

	return srvInstance, nil
}

func (srv *Server) Start() error {
	return srv.HttpServer.ListenAndServe()
}

func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.HttpServer.Shutdown(ctx)
}

// @Summary Get last rates
// @Description Retrieves the latest rates for specified cryptocurrencies.
// @Tags Coins
// @Accept json
// @Produce json
// @Param request body dto.RequestDTO true "Request containing coin titles" example(BTC,ETH)
// @Success 200 {object} dto.CoinDTO
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /rates/last [post]
func (srv *Server) getLastRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.RequestDTO
	if err := decodeRequest(r, &req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
	}
	req.Titles = removeEmptyStrings(req.Titles)

	if len(req.Titles) == 0 {
		sendErrorResponse(w, http.StatusBadRequest, "titles list cannot be empty")
		return
	}
	coins, err := srv.Service.GetLastRates(r.Context(), req.Titles)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]dto.CoinDTO, len(coins))
	for i, coin := range coins {
		dtos[i] = dto.CoinDTO{
			Title: coin.Title,
			Cost:  coin.Cost,
		}
	}
	responseDTO := dto.ResponseDTO{
		Coins: dtos,
	}

	jsonResponse(w, responseDTO)
}
func removeEmptyStrings(slices []string) []string {
	var result []string
	for _, str := range slices {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
func IsWrappedError(err error, target error) bool {
	for err != nil {
		if errors.Is(err, target) {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

// @Summary Get aggregate rates
// @Description Aggregates rates for specified cryptocurrencies based on given parameters.
// @Tags Coins
// @Accept json
// @Produce json
// @Param request body dto.RequestDTO true "Request containing coin titles and aggregation type"
// @Success 200 {object} dto.CoinDTO
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /rates/aggregate [post]
func (srv *Server) getAggregateRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.RequestDTO
	if err := decodeRequest(r, &req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Titles = removeEmptyStrings(req.Titles)
	if len(req.Titles) == 0 {
		sendErrorResponse(w, http.StatusBadRequest, "titles list cannot be empty")
		return
	}

	validTypes := map[string]bool{"MIN": true, "MAX": true, "AVG": true}
	if !validTypes[req.AggType] {
		sendErrorResponse(w, http.StatusBadRequest, "invalid aggregation type '"+req.AggType+"'")
		return
	}

	coins, err := srv.Service.GetAggregateRates(r.Context(), req.Titles, req.AggType)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidParam) {
			sendErrorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	dtos := make([]dto.CoinDTO, len(coins))
	for i, coin := range coins {
		dtos[i] = dto.CoinDTO{
			Title: coin.Title,
			Cost:  coin.Cost,
		}
	}

	responseDTO := dto.ResponseDTO{
		Coins: dtos,
	}

	jsonResponse(w, responseDTO)
}

func decodeRequest(r *http.Request, decReq any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return errors.New("empty request body")
	}

	v := validator.New()
	if err := json.NewDecoder(r.Body).Decode(decReq); err != nil {
		return err
	}
	if err := v.Struct(decReq); err != nil {
		return err
	}
	return nil
}

func sendErrorResponse(w http.ResponseWriter, status int, message string) {
	errorResp := dto.ErrorResponseDTO{
		Code:    status,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}
