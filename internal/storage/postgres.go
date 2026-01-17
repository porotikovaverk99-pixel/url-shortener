package storage

import (
	"context"

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
			short_id VARCHAR(50) PRIMARY KEY,
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
		"INSERT INTO urls (short_id, original_url) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		shortID, originalURL)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostgresStorage) Get(ctx context.Context, shortID string) (string, error) {
	var originalURL string
	err := ps.pool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE short_id = $1", shortID).Scan(&originalURL)
	if err != nil {
		return "", ErrURLNotFound
	}
	return originalURL, nil
}

func (ps *PostgresStorage) FindIDByURL(ctx context.Context, url string) (string, error) {
	var shortID string
	err := ps.pool.QueryRow(ctx,
		"SELECT short_id FROM urls WHERE original_url = $1", url).Scan(&shortID)
	if err != nil {
		return "", ErrIDNotFound
	}
	return shortID, nil
}

func (ps *PostgresStorage) Ping(ctx context.Context) error {
	return ps.pool.Ping(ctx)
}
