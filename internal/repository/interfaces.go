package repository

import (
	"context"
	"errors"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
)

var (
	ErrURLNotFound      = errors.New("URL not found")
	ErrIDNotFound       = errors.New("ID not found")
	ErrIDAlreadyExists  = errors.New("ID already exists")
	ErrURLAlreadyExists = errors.New("URL already exists")
)

type URLRepository interface {
	Save(ctx context.Context, short string, original string) error
	SaveBatch(ctx context.Context, batch []model.BatchItem) error
	Get(ctx context.Context, short string) (string, error)
	FindIDByURL(ctx context.Context, url string) (string, error)
	FindIDByURLs(ctx context.Context, urls []string) (map[string]string, error)
	Ping(ctx context.Context) error
	GetAll(ctx context.Context) (map[string]string, error)
}
