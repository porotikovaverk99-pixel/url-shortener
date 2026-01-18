package storage

import (
	"context"
	"errors"
)

var (
	ErrURLNotFound     = errors.New("URL not found")
	ErrIDNotFound      = errors.New("ID not found")
	ErrIDAlreadyExists = errors.New("URL already exists")
)

type BatchItem struct {
	ShortURL    string
	OriginalURL string
}

type URLStorage interface {
	Save(ctx context.Context, short string, original string) error
	SaveBatch(ctx context.Context, batch []BatchItem) error
	Get(ctx context.Context, short string) (string, error)
	FindIDByURL(ctx context.Context, url string) (string, error)
	FindIDByURLs(ctx context.Context, urls []string) (map[string]string, error)
	Ping(ctx context.Context) error
}
