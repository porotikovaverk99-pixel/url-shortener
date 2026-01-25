package main

import (
	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/gzip"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/handlers"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/logger"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/server"
	strg "github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"go.uber.org/zap"
)

func main() {

	cfg := config.ParseFlags()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic("Error ocurs while initializing logger: " + err.Error())
	}

	var urlStorage strg.URLStorage
	var err error

	if cfg.DatabaseDSN != "" {
		urlStorage, err = strg.NewPostgresStorage(cfg.DatabaseDSN)
		if err != nil {
			panic("Error occurs while initializing PostgreSQL storage: " + err.Error())
		}
		logger.Log.Info("Using PostgreSQL storage")
	} else {
		urlStorage, err = strg.NewMemoryStorage(cfg.FileStoragePath)
		if err != nil {
			panic("Error occurs while initializing file storage: " + err.Error())
		}
		logger.Log.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
	}

	server := server.New(cfg.RunAddr)

	handlerMain := logger.RequestLogger(gzip.GzipMiddleware(handlers.URLHandler(urlStorage, cfg.BaseURL)))
	handlerShorten := logger.RequestLogger(gzip.GzipMiddleware(handlers.URLHandlerShorten(urlStorage, cfg.BaseURL)))
	handlerPing := logger.RequestLogger(gzip.GzipMiddleware(handlers.URLHandlerPing(urlStorage)))
	handlerShortenBatch := logger.RequestLogger(gzip.GzipMiddleware(handlers.URLHandlerShortenBatch(urlStorage, cfg.BaseURL)))
	handlerGetAll := logger.RequestLogger(gzip.GzipMiddleware(handlers.URLGetAll(urlStorage)))

	server.HandleFunc("/", handlerMain.ServeHTTP)
	server.HandleFunc("/{id}", handlerMain.ServeHTTP)
	server.HandleFunc("/api/shorten", handlerShorten.ServeHTTP)
	server.HandleFunc("/ping", handlerPing.ServeHTTP)
	server.HandleFunc("/api/shorten/batch", handlerShortenBatch.ServeHTTP)
	server.HandleFunc("/api/getAll", handlerGetAll.ServeHTTP)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))

	err = server.Run()
	if err != nil {
		panic("Error occurs while running server: " + err.Error())
	}

}
