package storage

import (
	"context"

	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	err = runMigrations(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &PostgresStorage{pool: pool}, nil
}

func runMigrations(dsn string) error {

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	migrationsPath := filepath.Join(exeDir, "migrations")

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = "migrations"
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			return fmt.Errorf("migrations directory not found: %w", err)
		}
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
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
				case "idx_urls_original_url":
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

func (ps *PostgresStorage) GetAll(ctx context.Context) (map[string]string, error) {

	rows, err := ps.pool.Query(ctx, "SELECT short_url, original_url FROM urls")

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
