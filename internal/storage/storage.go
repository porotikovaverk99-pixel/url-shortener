package storage

import (
	"errors"
)

type MemoryStorage struct {
	data map[string]string
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
	if _, ok := ms.data[short]; ok {
		return errors.New("URL already exists")
	}
	ms.data[short] = original
	return nil
}

func (ms *MemoryStorage) Get(short string) (string, error) {
	if original, ok := ms.data[short]; ok {
		return original, nil
	}
	return "", errors.New("URL not found")
}
