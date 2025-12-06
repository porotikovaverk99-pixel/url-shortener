package storage

import (
	"errors"
	"sync"
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
	Save(short string, original string) error
	Get(short string) (string, error)
}

func (ms *MemoryStorage) Save(short string, original string) error {
	ms.mu.Lock() 
    defer ms.mu.Unlock()
	if _, ok := ms.data[short]; ok {
		return errors.New("URL already exists")
	}
	ms.data[short] = original
	return nil
}

func (ms *MemoryStorage) Get(short string) (string, error) {
	ms.mu.RLock()
    defer ms.mu.RUnlock()
	if original, ok := ms.data[short]; ok {
		return original, nil
	}
	return "", errors.New("URL not found")
}
