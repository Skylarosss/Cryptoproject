package storage

import (
	"context"
	"database/sql"
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
		return nil, errors.Wrapf(entities.ErrInternal, "New pool failure: %v", err)
	}
	st.dbPool = pool
	return st, nil
}
func (s *Storage) Close() {
	s.once.Do(
		func() {
			s.cancel()
			s.dbPool.Close()
		})
}

func (s *Storage) Store(ctx context.Context, coins []*entities.Coin) error {
	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errors.Wrapf(entities.ErrInternal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	sqlQuery := `
    INSERT INTO coins (title, cost)
    VALUES ($1, $2)
    ON CONFLICT (title) DO UPDATE SET cost = EXCLUDED.cost
    `
	//без цикла
	for _, coin := range coins {
		_, err = tx.Exec(ctx, sqlQuery, coin.Title, coin.Cost)
		if err != nil {
			return errors.Wrapf(entities.ErrInternal, "failed to insert/update coin with title=%q", coin.Title)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(entities.ErrInternal, "failed to commit transaction")
	}

	return nil
}

func (s *Storage) GetCoinsList(ctx context.Context) ([]string, error) {
	rows, err := s.dbPool.Query(ctx, "SELECT DISTINCT title FROM coins ORDER BY title ASC")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch distinct titles of coins")
	}
	defer rows.Close()

	titles := make([]string, 0)
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, errors.Wrap(err, "failed to scan row into title")
		}
		titles = append(titles, title)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error occurred while iterating over results")
	}

	return titles, nil
}

func (s *Storage) GetActualCoins(ctx context.Context, titles []string) ([]entities.Coin, error) {
	query := `
    SELECT title, cost
    FROM coins
    WHERE title = ANY($1::TEXT[])
    ORDER BY title ASC
    `

	rows, err := s.dbPool.Query(ctx, query, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute select query for actual coins")
	}
	defer rows.Close()

	result := make([]entities.Coin, 0)
	for rows.Next() {
		var coin entities.Coin
		if err := rows.Scan(&coin.Title, &coin.Cost); err != nil {
			return nil, errors.Wrap(err, "failed to scan row into coin object")
		}
		result = append(result, coin)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error occurred while iterating over results")
	}

	return result, nil
}

// убрать дубликат
func (s *Storage) GetAggregateCoins(ctx context.Context, titles []string, aggType string) ([]entities.Coin, error) {
	switch aggType {
	case "AVG":
		query := `
        SELECT title, AVG(cost) AS avg_cost
        FROM coins
        WHERE title = ANY($1::TEXT[])
        GROUP BY title
        ORDER BY title ASC
        `
		return s.queryAggregatedCoins(ctx, query, titles)
	case "MIN":
		query := `
        SELECT title, MIN(cost) AS min_cost
        FROM coins
        WHERE title = ANY($1::TEXT[])
        GROUP BY title
        ORDER BY title ASC
        `
		return s.queryAggregatedCoins(ctx, query, titles)
	case "MAX":
		query := `
        SELECT title, MAX(cost) AS max_cost
        FROM coins
        WHERE title = ANY($1::TEXT[])
        GROUP BY title
        ORDER BY title ASC
        `
		return s.queryAggregatedCoins(ctx, query, titles)
	default:
		return nil, errors.Wrap(entities.ErrInvalidParam, "unsupported aggregation type")
	}
}

func (s *Storage) queryAggregatedCoins(ctx context.Context, query string, titles []string) ([]entities.Coin, error) {
	rows, err := s.dbPool.Query(ctx, query, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute aggregated query")
	}
	defer rows.Close()

	result := make([]entities.Coin, 0)
	for rows.Next() {
		var coin entities.Coin
		var cost sql.NullFloat64
		if err := rows.Scan(&coin.Title, &cost); err != nil {
			return nil, errors.Wrap(err, "failed to scan row into coin object")
		}
		if cost.Valid {
			coin.Cost = cost.Float64
		}
		result = append(result, coin)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error occurred while iterating over results")
	}

	return result, nil
}
