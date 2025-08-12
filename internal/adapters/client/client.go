package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"Cryptoproject/internal/entities"

	"log/slog"

	"github.com/pkg/errors"
)

const (
	baseURL         = "https://min-api.cryptocompare.com/data/"
	priceMulti      = "pricemulti"
	defaultCurrency = "USD"
	fsymsQuery      = "fsyms"
	tsymsQuery      = "tsyms"
)

type Client struct {
	httpClient *http.Client
	costIn     string
}

type ClientOption func(*Client)

func WithCustomCostIn(costIn string) ClientOption {
	return func(c *Client) {
		c.costIn = costIn
	}
}

func (c *Client) setOption(opts ...ClientOption) {
	for _, opt := range opts {
		opt(c)
	}
}

func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
		costIn:     defaultCurrency,
	}

	c.setOption(opts...)

	if c.costIn == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "CostIn cannot be empty")
	}

	slog.Info("Client initialized", "cost_in", c.costIn)

	return c, nil
}
func (c *Client) GetActualRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	slog.Info("Fetching actual coin rates", "titles", titles)

	u, err := url.Parse(fmt.Sprintf("%s%s", baseURL, priceMulti))
	if err != nil {
		slog.Error("Failed to parse URL", "err", err)
		return nil, errors.Wrap(err, "failed to parse URL")
	}

	q := u.Query()
	q.Set(fsymsQuery, strings.Join(titles, ","))
	q.Set(tsymsQuery, c.costIn)

	u.RawQuery = q.Encode()
	requestUrl := u.String()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		slog.Error("Failed to build request", "err", err)
		return nil, errors.Wrap(err, "failed to build request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("HTTP request failed", "err", err)
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = errors.Wrapf(err, "failed to close response body: %v", cerr)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "err", err)
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode >= http.StatusBadRequest {
		message := string(bodyBytes)
		slog.Error("API returned an error", "status_code", resp.StatusCode, "response_body", message)
		return nil, errors.Errorf("API returned an error (%d): %s", resp.StatusCode, message)
	}

	var result map[string]map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		slog.Error("Couldn't parse response body", "err", err)
		return nil, errors.Wrap(err, "couldn't parse response body")
	}

	coinResults := make([]entities.Coin, 0, len(result))
	for title, priceMap := range result {
		cost, exists := priceMap[c.costIn]
		if !exists {
			slog.Info("Price data missing for currency", "currency", title)
			continue
		}

		floatCost, ok := cost.(float64)
		if !ok {
			slog.Error("Unexpected format of price data", "data", cost)
			continue
		}

		coin, err := entities.NewCoin(title, floatCost)
		if err != nil {
			slog.Error("Failed to create coin entity", "err", err)
			return nil, errors.Wrap(err, "failed to create coin entity")
		}
		coinResults = append(coinResults, *coin)
	}

	slog.Info("Fetched coin rates successfully", "number_of_coins", len(coinResults))

	return coinResults, nil
}
