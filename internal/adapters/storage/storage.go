package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"Cryptoproject/internal/entities"
)

type Storage struct {
	dbPool *pgxpool.Pool
	cancel context.CancelFunc
	once   sync.Once
}

func NewStorage(connStr string) (*Storage, error) {
	if connStr == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "connection string is empty")
	}
	st := &Storage{}
	ctx, cancel := context.WithCancel(context.Background())
	st.cancel = cancel
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		slog.Error("Failed to connect to database", "conn_str", connStr, "err", err)
		return nil, errors.Wrapf(entities.ErrInternal, "New pool failure: %v", err)
	}
	st.dbPool = pool
	slog.Info("Database connection established", "conn_str", connStr)
	return st, nil
}

func (s *Storage) Close() {
	s.once.Do(
		func() {
			s.cancel()
			s.dbPool.Close()
			slog.Info("Database connection closed")
		},
	)
}

func (s *Storage) Store(ctx context.Context, coins []entities.Coin) error {
	data := make([][]interface{}, len(coins))
	for i, coin := range coins {
		data[i] = []interface{}{coin.Title, coin.Cost}
	}

	_, err := s.dbPool.CopyFrom(ctx, pgx.Identifier{"coins"}, []string{"title", "cost"}, pgx.CopyFromRows(data))
	if err != nil {
		slog.Error("Failed to perform bulk insert using COPY", "err", err)
		return errors.Wrapf(entities.ErrInternal, "failed to perform bulk insert using COPY: %v", err)
	}

	slog.Info("Bulk insert completed successfully", "number_of_coins", len(coins))
	return nil
}

func (s *Storage) GetCoinsList(ctx context.Context) ([]string, error) {
	rows, err := s.dbPool.Query(ctx, "SELECT DISTINCT title FROM coins ORDER BY title ASC")
	if err != nil {
		slog.Error("Failed to fetch distinct titles of coins", "err", err)
		return nil, errors.Wrap(err, "failed to fetch distinct titles of coins")
	}
	defer rows.Close()

	titles := make([]string, 0)
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			slog.Error("Failed to scan row into title", "err", err)
			return nil, errors.Wrap(err, "failed to scan row into title")
		}
		titles = append(titles, title)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error occurred while iterating over results", "err", err)
		return nil, errors.Wrap(err, "error occurred while iterating over results")
	}

	slog.Info("Distinct coin titles fetched successfully", "number_of_titles", len(titles))
	return titles, nil
}

func (s *Storage) GetActualCoins(ctx context.Context, titles []string) ([]entities.Coin, error) {
	query := `
    SELECT title, cost
        FROM coins
        WHERE title = ANY($1::TEXT[]) AND actual_at IN (
            SELECT MAX(actual_at) 
            FROM coins 
            WHERE title = ANY($1::TEXT[])
            GROUP BY title
        )
        ORDER BY title ASC
    `

	rows, err := s.dbPool.Query(ctx, query, titles)
	if err != nil {
		slog.Error("Failed to execute select query for actual coins", "err", err)
		return nil, errors.Wrap(err, "failed to execute select query for actual coins")
	}
	defer rows.Close()

	result := make([]entities.Coin, 0)
	for rows.Next() {
		var coin entities.Coin
		if err := rows.Scan(&coin.Title, &coin.Cost); err != nil {
			slog.Error("Failed to scan row into coin object", "err", err)
			return nil, errors.Wrap(err, "failed to scan row into coin object")
		}
		result = append(result, coin)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error occurred while iterating over results", "err", err)
		return nil, errors.Wrap(err, "error occurred while iterating over results")
	}

	slog.Info("Actual coin rates fetched successfully", "number_of_coins", len(result))
	return result, nil
}

func (s *Storage) GetAggregateCoins(ctx context.Context, titles []string, aggType string) ([]entities.Coin, error) {
	var aggFunc string

	switch aggType {
	case "AVG":
		aggFunc = "AVG(cost)"
	case "MIN":
		aggFunc = "MIN(cost)"
	case "MAX":
		aggFunc = "MAX(cost)"
	default:
		return nil, errors.Wrap(entities.ErrInvalidParam, "unsupported aggregation type")
	}

	query := fmt.Sprintf(`
        SELECT title, %s AS cost
        FROM coins
        WHERE title = ANY($1::TEXT[])
        GROUP BY title
        ORDER BY title ASC
    `, aggFunc)

	rows, err := s.dbPool.Query(ctx, query, titles)
	if err != nil {
		slog.Error("Failed to execute aggregated query", "err", err)
		return nil, errors.Wrap(err, "failed to execute aggregated query")
	}
	defer rows.Close()

	result := make([]entities.Coin, 0)
	for rows.Next() {
		var coin entities.Coin
		var cost sql.NullFloat64
		if err := rows.Scan(&coin.Title, &cost); err != nil {
			slog.Error("Failed to scan row into coin object", "err", err)
			return nil, errors.Wrap(err, "failed to scan row into coin object")
		}
		if cost.Valid {
			coin.Cost = cost.Float64
		}
		result = append(result, coin)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error occurred while iterating over results", "err", err)
		return nil, errors.Wrap(err, "error occurred while iterating over results")
	}

	slog.Info("Aggregated coin rates fetched successfully", "number_of_coins", len(result))
	return result, nil
}
