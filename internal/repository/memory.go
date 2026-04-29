package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
)

type MemoryStorage struct {
	data     map[string]string
	mu       sync.RWMutex
	filePath string
}

func NewMemoryStorage(filePath string) (*MemoryStorage, error) {
	ms := &MemoryStorage{
		data:     make(map[string]string),
		filePath: filePath,
	}
	if err := ms.ReadFromFile(); err != nil {
		return nil, err
	}
	return ms, nil
}

func (ms *MemoryStorage) Save(ctx context.Context, short string, original string, userID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if _, ok := ms.data[short]; ok {
		return ErrIDAlreadyExists
	}

	ms.data[short] = original

	dataCopy := make(map[string]string)
	for k, v := range ms.data {
		dataCopy[k] = v
	}

	return ms.WriteToFile(dataCopy)
}

func (ms *MemoryStorage) SaveBatch(ctx context.Context, batch []model.BatchItem, userID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, item := range batch {
		ms.data[item.ShortURL] = item.OriginalURL
	}

	dataCopy := make(map[string]string)
	for k, v := range ms.data {
		dataCopy[k] = v
	}

	return ms.WriteToFile(dataCopy)
}

func (ms *MemoryStorage) Get(ctx context.Context, short string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if original, ok := ms.data[short]; ok {
		return original, nil
	}
	return "", ErrURLNotFound
}

func (ms *MemoryStorage) FindIDByURL(ctx context.Context, url string, userID string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for k, v := range ms.data {
		if v == url {
			return k, nil
		}
	}
	return "", ErrIDNotFound
}

func (ms *MemoryStorage) FindIDByURLs(ctx context.Context, urls []string, userID string) (map[string]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make(map[string]string)
	for shortID, originalURL := range ms.data {
		for _, url := range urls {
			if originalURL == url {
				result[originalURL] = shortID
				break
			}
		}
	}
	return result, nil
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

type FileData struct {
	ID          string `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (ms *MemoryStorage) ReadFromFile() error {

	if _, err := os.Stat(ms.filePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(ms.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return nil
	}

	var fileData []FileData
	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&fileData); err != nil {
		return err
	}

	for _, el := range fileData {
		ms.data[el.ShortURL] = el.OriginalURL
	}

	return nil

}

func (ms *MemoryStorage) WriteToFile(data map[string]string) error {

	file, err := os.OpenFile(ms.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var fileData []FileData
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	i := 1

	for k, v := range data {
		fileData = append(fileData, FileData{
			ID:          fmt.Sprintf("%d", i),
			ShortURL:    k,
			OriginalURL: v,
		})
		i++
	}

	return encoder.Encode(fileData)

}

func (ms *MemoryStorage) GetUserURLs(ctx context.Context, userID string) (map[string]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.data, nil
}

func (ms *MemoryStorage) CreateUser(ctx context.Context, userID string) error {
	return nil
}

func (ms *MemoryStorage) MarkURLsAsDeleted(ctx context.Context, urls []string, userID string) error {
	return nil
}

func (ms *MemoryStorage) Close() error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.filePath != "" {
		return ms.WriteToFile(ms.data)
	}
	return nil
}
