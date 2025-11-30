package main

import (
	"url-shortener/internal/handlers"
	"url-shortener/internal/storage"
	"url-shortener/internal/server"
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
