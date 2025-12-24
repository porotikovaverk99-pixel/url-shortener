package storage

import (
	"errors"
	"sync"
	"context"
)

var (
    ErrURLNotFound      = errors.New("URL not found")
	ErrIDNotFound      = errors.New("ID not found")
    ErrIDAlreadyExists = errors.New("URL already exists")
)

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

type URLStorage interface {
	Save(ctx context.Context, short string, original string) error
	Get(ctx context.Context, short string) (string, error)
	FindIDByURL(ctx context.Context, url string) (string, error)
}

func (ms *MemoryStorage) Save(ctx context.Context, short string, original string) error {
	ms.mu.Lock() 
    defer ms.mu.Unlock()
	if _, ok := ms.data[short]; ok {
		return ErrIDAlreadyExists
	}
	ms.data[short] = original
	return nil
}

func (ms *MemoryStorage) Get(ctx context.Context, short string) (string, error) {
	ms.mu.RLock()
    defer ms.mu.RUnlock()
	if original, ok := ms.data[short]; ok {
		return original, nil
	}
	return "", ErrURLNotFound
}

func (ms *MemoryStorage) FindIDByURL(ctx context.Context, url string) (string, error) {
	ms.mu.RLock()
    defer ms.mu.RUnlock()
	for k, v := range ms.data {
		if v == url {
			return k, nil
		}
	}
	return "", ErrIDNotFound
}
