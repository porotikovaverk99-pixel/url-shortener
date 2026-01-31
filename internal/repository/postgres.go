package repository

import (
	"context"

	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
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

	needMigrations, err := checkIfMigrationsNeeded(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to check migrations status: %w", err)
	}

	if needMigrations {
		err = runMigrations(dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	return &PostgresStorage{pool: pool}, nil
}

func checkIfMigrationsNeeded(pool *pgxpool.Pool) (bool, error) {

	var schemaExists bool
	err := pool.QueryRow(context.Background(),
		`SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'schema_migrations'
        )`).Scan(&schemaExists)

	if err != nil {
		return false, err
	}

	if !schemaExists {
		return true, nil
	}

	var currentVersion uint
	var dirty bool
	err = pool.QueryRow(context.Background(),
		`SELECT version, dirty FROM schema_migrations`).Scan(&currentVersion, &dirty)

	if err != nil {
		return false, err
	}

	if dirty {
		return false, fmt.Errorf("migrations are in dirty state")
	}

	return false, nil
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
	resTag, err := ps.pool.Exec(ctx,
		`INSERT INTO urls (short_url, original_url) 
         VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		shortID, originalURL)

	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}

	if resTag.RowsAffected() == 0 {
		return ErrIDAlreadyExists
	}

	return nil
}

func (ps *PostgresStorage) SaveBatch(ctx context.Context, batch []model.BatchItem) error {

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
		resTag, err := tx.Exec(ctx,
			`INSERT INTO urls (short_url, original_url) 
         	VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			item.ShortURL, item.OriginalURL)

		if err != nil {
			return fmt.Errorf("failed to save URL: %w", err)
		}

		if resTag.RowsAffected() == 0 {
			return ErrIDAlreadyExists
		}

	}

	return tx.Commit(ctx)
}

func (ps *PostgresStorage) Get(ctx context.Context, shortID string) (string, error) {
	var originalURL string
	err := ps.pool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE short_url = $1", shortID).Scan(&originalURL)
	if err != nil {
		return "", ErrURLNotFound
	}
	return originalURL, nil
}

func (ps *PostgresStorage) FindIDByURL(ctx context.Context, url string) (string, error) {
	var shortURL string
	err := ps.pool.QueryRow(ctx,
		"SELECT short_url FROM urls WHERE original_url = $1", url).Scan(&shortURL)
	if err != nil {
		return "", ErrIDNotFound
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
