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

	"github.com/pkg/errors"
)

const (
	baseURL         = "https://min-api.cryptocompare.com/data/" // Базовый URL API
	priceMulti      = "pricemulti"                              // Эндпоинт для получения мультивалютных курсов
	defaultCurrency = "USD"                                     // Валюта по умолчанию для отображения цен
	fsymsQuery      = "fsyms"                                   // Параметр запроса для обозначения валют
	tsymsQuery      = "tsyms"                                   // Параметр запроса для обозначения целевой валюты
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
	return c, nil
}

func (c *Client) GetActualRates(ctx context.Context, titles []string) ([]entities.Coin, error) {

	u, err := url.Parse(fmt.Sprintf("%s%s", baseURL, priceMulti))
	if err != nil {
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
		return nil, errors.Wrap(err, "failed to build request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Wrap(err, "API returned status code")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	var result map[string]map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse response body")
	}

	coinResults := make([]entities.Coin, 0, len(result))
	for title, priceMap := range result {
		cost, exists := priceMap[c.costIn].(float64)
		if !exists {
			continue
		}
		coin, err := entities.NewCoin(title, cost)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create coin entity")
		}
		coinResults = append(coinResults, *coin)
	}

	return coinResults, nil
}
