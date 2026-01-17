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

type URLStorage interface {
	Save(ctx context.Context, short string, original string) error
	Get(ctx context.Context, short string) (string, error)
	FindIDByURL(ctx context.Context, url string) (string, error)
	Ping(ctx context.Context) error
}
