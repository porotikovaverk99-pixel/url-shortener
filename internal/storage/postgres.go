package storage

import (
	"context"

	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
    		short_url VARCHAR(50) UNIQUE NOT NULL,
    		original_url TEXT NOT NULL,
    		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{pool: pool}, nil
}

func (ps *PostgresStorage) Save(ctx context.Context, shortID, originalURL string) error {
	_, err := ps.pool.Exec(ctx,
		`INSERT INTO urls (short_url, original_url) 
         VALUES ($1, $2)`,
		shortID, originalURL)

	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			if pgErr.Code == pgerrcode.UniqueViolation {
				switch pgErr.ConstraintName {
				case "urls_short_url_key":
					return fmt.Errorf("failed to save URL: %w", ErrIDAlreadyExists)
				case "idx_urls_original_url_unique":
					return fmt.Errorf("failed to save URL: %w", ErrURLAlreadyExists)
				}
			}
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}

	return nil
}

func (ps *PostgresStorage) SaveBatch(ctx context.Context, batch []BatchItem) error {

	conn, err := ps.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	for _, item := range batch {
		cmdTag, err := tx.Exec(ctx,
			`INSERT INTO urls (short_url, original_url) 
         	VALUES ($1, $2) 
         	ON CONFLICT (short_url) DO NOTHING`,
			item.ShortURL, item.OriginalURL)

		if err != nil {
			return err
		}

		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("%w", ErrIDAlreadyExists)
		}

	}

	return tx.Commit(ctx)
}

func (ps *PostgresStorage) Get(ctx context.Context, shortID string) (string, error) {
	var originalURL string
	err := ps.pool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE short_url = $1", shortID).Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("%w", ErrURLNotFound)
	}
	return originalURL, nil
}

func (ps *PostgresStorage) FindIDByURL(ctx context.Context, url string) (string, error) {
	var shortURL string
	err := ps.pool.QueryRow(ctx,
		"SELECT short_url FROM urls WHERE original_url = $1", url).Scan(&shortURL)
	if err != nil {
		return "", fmt.Errorf("%w", ErrIDNotFound)
	}
	return shortURL, nil
}

func (ps *PostgresStorage) FindIDByURLs(ctx context.Context, urls []string) (map[string]string, error) {

	if len(urls) == 0 {
		return map[string]string{}, nil
	}

	rows, err := ps.pool.Query(ctx, "SELECT short_url, original_url FROM urls WHERE original_url = ANY($1)", urls)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var shortURL, originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return nil, err
		}
		result[originalURL] = shortURL
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (ps *PostgresStorage) Ping(ctx context.Context) error {
	return ps.pool.Ping(ctx)
}
