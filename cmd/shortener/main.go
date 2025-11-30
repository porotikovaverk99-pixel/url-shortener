package main

import (
	"github.com/porotikovaverk99-pixel/url-shortener/internal/handlers"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/server"
)

func main() {
	storage := storage.NewMemoryStorage()
	handler := handlers.URLHandler(storage)
	server := server.New(handler)
	err := server.Run()
	if err != nil {
		panic(err)
	}
}
