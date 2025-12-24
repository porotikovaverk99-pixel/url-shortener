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
	
	server := server.New(cfg.RunAddr)

	handlerMain := logger.RequestLogger(handlers.URLHandler(storage, cfg.BaseURL))
	handlerShorten := logger.RequestLogger(handlers.URLHandlerShorten(storage, cfg.BaseURL))

	server.RegisterHandler("/", handlerMain)
	server.RegisterHandler("/{id}", handlerMain)
	server.RegisterHandler("/api/shorten", handlerShorten)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))
	 
	err := server.Run() 
	if err != nil {
		panic("Error occurs while running server: " + err.Error())
	} 

}
