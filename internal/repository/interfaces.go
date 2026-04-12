package repository

import (
	"context"
	"errors"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
)

var (
	// ErrURLNotFound возвращается когда URL не найден в хранилище
	ErrURLNotFound = errors.New("URL not found")
	// ErrIDNotFound возвращается когда ID не найден в хранилище
	ErrIDNotFound = errors.New("ID not found")
	// ErrIDAlreadyExists возвращается при попытке сохранить уже существующий ID
	ErrIDAlreadyExists = errors.New("ID already exists")
	// ErrURLAlreadyExists возвращается при попытке сохранить уже существующий URL
	ErrURLAlreadyExists = errors.New("URL already exists")
	// ErrURLDeleted возвращается когда URL был помечен как удаленный
	ErrURLDeleted = errors.New("URL deleted")
)

// URLRepository определяет интерфейс для работы с хранилищем URL.
type URLRepository interface {
	// Save сохраняет короткий URL и оригинальный URL для пользователя.
	Save(ctx context.Context, short string, original string, userID string) error

	// SaveBatch сохраняет пакет URL для пользователя.
	SaveBatch(ctx context.Context, batch []model.BatchItem, userID string) error

	// Get возвращает оригинальный URL по короткому идентификатору.
	Get(ctx context.Context, short string) (string, error)

	// FindIDByURL возвращает короткий ID по оригинальному URL для пользователя.
	FindIDByURL(ctx context.Context, url string, userID string) (string, error)

	// FindIDByURLs возвращает map соответствий оригинальных URL их коротким ID для пользователя.
	FindIDByURLs(ctx context.Context, urls []string, userID string) (map[string]string, error)

	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error

	// GetUserURLs возвращает все URL принадлежащие пользователю.
	GetUserURLs(ctx context.Context, userID string) (map[string]string, error)

	// MarkURLsAsDeleted помечает URL как удаленные для пользователя.
	MarkURLsAsDeleted(ctx context.Context, urls []string, userID string) error

	// Close закрывает соединение с хранилищем.
	Close() error
}

// UserRepository определяет интерфейс для работы с пользователями.
type UserRepository interface {
	// CreateUser создает нового пользователя.
	CreateUser(ctx context.Context, userID string) error
}
