package main

import (

	"github.com/porotikovaverk99-pixel/url-shortener/internal/handlers"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/server"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/logger"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/gzip"
	"go.uber.org/zap"

)

func main() {

	cfg := config.ParseFlags()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic("Error ocurs while initializing logger: " + err.Error())
	}

	storage := storage.NewMemoryStorage()
	
	server := server.New(cfg.RunAddr)

	handlerMain := logger.RequestLogger(gzip.GzipMiddleware(handlers.URLHandler(storage, cfg.BaseURL)))
	handlerShorten := logger.RequestLogger(gzip.GzipMiddleware(handlers.URLHandlerShorten(storage, cfg.BaseURL)))

	server.HandleFunc("/", handlerMain.ServeHTTP)
	server.HandleFunc("/{id}", handlerMain.ServeHTTP)
	server.Post("/api/shorten", handlerShorten.ServeHTTP)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))
	 
	err := server.Run() 
	if err != nil {
		panic("Error occurs while running server: " + err.Error())
	} 

}
