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
	Save(ctx context.Context, short string, original string, userID string) error
	SaveBatch(ctx context.Context, batch []model.BatchItem, userID string) error
	Get(ctx context.Context, short string) (string, error)
	FindIDByURL(ctx context.Context, url string, userID string) (string, error)
	FindIDByURLs(ctx context.Context, urls []string, userID string) (map[string]string, error)
	Ping(ctx context.Context) error
	GetUserURLs(ctx context.Context, userID string) (map[string]string, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, userID string) error
}
