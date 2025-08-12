package http

import (
	"context"
	"encoding/json"
	"log/slog"
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

	// Документация OpenAPI
	// @Title Cryptocurrency Rates API
	// @Version 1.0
	// @Description This API provides cryptocurrency rate information.
	// @Host localhost:8080
	// @BasePath /
	// @schemes http
	// @produces json
	// @consumes json
	router.Get("/ping", srvInstance.pingHandler)
	srvInstance.Router.Post("/rates/last", srvInstance.getLastRates)
	srvInstance.Router.Post("/rates/aggregate", srvInstance.getAggregateRates)

	return srvInstance, nil
}
func (srv *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "OK"}`))
}
func (srv *Server) Start() error {
	slog.Info("HTTP server started", "addr", srv.HttpServer.Addr)
	return srv.HttpServer.ListenAndServe()
}

func (srv *Server) Shutdown(ctx context.Context) error {
	slog.Info("HTTP server shutting down")
	return srv.HttpServer.Shutdown(ctx)
}

// @Summary Get last rates
// @Description Retrieves the latest rates for specified cryptocurrencies.
// @Tags Coins
// @Accept json
// @Produce json
// @Param request body dto.RequestDTO true "Request containing coin titles" example(BTC,ETH)
// @Success 200 {object} dto.ResponseDTO
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /rates/last [post]
func (srv *Server) getLastRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.RequestDTO
	if err := srv.decodeRequest(r, &req); err != nil {
		slog.Error("Failed to decode request", "err", err)
		srv.errProcessing(w, err)
		return
	}
	req.Titles = srv.removeEmptyStrings(req.Titles)

	if len(req.Titles) == 0 {
		slog.Error("Empty titles list provided")
		srv.errProcessing(w, errors.Wrap(entities.ErrInvalidParam, "titles list cannot be empty"))
		return
	}

	coins, err := srv.Service.GetLastRates(r.Context(), req.Titles)
	if err != nil {
		slog.Error("Failed to get last rates", "err", err)
		srv.errProcessing(w, err)
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

	srv.jsonResponse(w, responseDTO)
	slog.Info("Successfully retrieved last rates", "number_of_coins", len(dtos))
}

// @Summary Get aggregate rates
// @Description Aggregates rates for specified cryptocurrencies based on given parameters.
// @Tags Coins
// @Accept json
// @Produce json
// @Param request body dto.RequestDTO true "Request containing coin titles and aggregation type"
// @Success 200 {object} dto.ResponseDTO
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /rates/aggregate [post]
func (srv *Server) getAggregateRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dto.RequestDTO
	if err := srv.decodeRequest(r, &req); err != nil {
		slog.Error("Failed to decode request", "err", err)
		srv.errProcessing(w, err)
		return
	}
	req.Titles = srv.removeEmptyStrings(req.Titles)
	if len(req.Titles) == 0 {
		slog.Error("Empty titles list provided")
		srv.errProcessing(w, errors.Wrap(entities.ErrInvalidParam, "titles list cannot be empty"))
		return
	}

	validTypes := map[string]bool{"MIN": true, "MAX": true, "AVG": true}
	if !validTypes[req.AggType] {
		slog.Error("Unsupported aggregation type", "agg_type", req.AggType)
		srv.errProcessing(w, errors.Wrap(entities.ErrInvalidParam, "invalid aggregation type"))
		return
	}

	coins, err := srv.Service.GetAggregateRates(r.Context(), req.Titles, req.AggType)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidParam) {
			slog.Error("Invalid parameter provided", "err", err)
			srv.errProcessing(w, err)
		} else if errors.Is(err, entities.ErrInternal) {
			slog.Error("Internal service error", "err", err)
			srv.errProcessing(w, err)
		} else {
			slog.Error("Failed to get aggregate rates", "err", err)
			srv.errProcessing(w, err)
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

	srv.jsonResponse(w, responseDTO)
	slog.Info("Successfully retrieved aggregate rates", "number_of_coins", len(dtos))
}

func (srv *Server) decodeRequest(r *http.Request, decReq any) error {
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

func (srv *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Unable to encode response", "err", err)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}

func (srv *Server) removeEmptyStrings(slices []string) []string {
	var result []string
	for _, str := range slices {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
func (srv *Server) errProcessing(resp http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	errDTO := dto.ErrorResponseDTO{
		Code:    statusCode,
		Message: err.Error(),
	}

	switch {
	case errors.Is(err, entities.ErrInvalidParam):
		errDTO.Code = http.StatusBadRequest
	case errors.Is(err, entities.ErrInternal):
		errDTO.Code = http.StatusForbidden
	case errors.Is(err, entities.ErrNotFound):
		errDTO.Code = http.StatusNotFound
	default:
		errors.Is(err, entities.ErrNotFound)
		errDTO.Code = http.StatusNotFound
		errDTO.Message = "coin does not exist or was not found in the provider"
	}

	errDtoData, err := json.Marshal(&errDTO)
	if err != nil {
		err := errors.Wrapf(entities.ErrInternal, "marshal failure: %v", err)
		slog.Error(err.Error())
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(errDTO.Code)
	resp.Write(errDtoData) //nolint:errcheck,gosec //ok
}
