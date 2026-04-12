package audit

import (
	"encoding/json"
	"os"
	"sync"
)

type FileObserver struct {
	path string
	mu   sync.Mutex
}

func NewFileObserver(path string) *FileObserver {
	return &FileObserver{path: path}
}

func (f *FileObserver) Send(event Event) error {

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(data) + "\n")

	return err
}
