package main

import (

	"github.com/porotikovaverk99-pixel/url-shortener/internal/handlers"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/server"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/logger"
	"go.uber.org/zap"

)

func main() {

	cfg := config.ParseFlags()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic("Error ocurs while initializing logger: " + err.Error())
	}

	storage := storage.NewMemoryStorage()
	handler := logger.RequestLogger(handlers.URLHandler(storage, cfg.BaseURL))
	server := server.New(handler, cfg.RunAddr)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))

	err := server.Run() 
	if err != nil {
		panic("Error occurs while running server: " + err.Error())
	} 

}
