package main

import (
	"math/rand"
    "time"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/handlers"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/server"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg := config.ParseFlags()

	storage := storage.NewMemoryStorage()
	handler := handlers.URLHandler(storage, cfg.BaseURL)
	server := server.New(handler, cfg.RunAddr)

	err := server.Run() 
	if err != nil {
		panic(err)
	}
}
