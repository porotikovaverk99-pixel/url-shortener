package main

import (
	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
	hdlr "github.com/porotikovaverk99-pixel/url-shortener/internal/handler"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/logger"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/server"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/service"
	"go.uber.org/zap"
)

func main() {

	cfg := config.ParseFlags()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic("Error ocurs while initializing logger: " + err.Error())
	}

	var URLRepository repository.URLRepository
	var err error

	if cfg.DatabaseDSN != "" {
		URLRepository, err = repository.NewPostgresStorage(cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Error occurs while initializing PostgreSQL storage", zap.Error(err))
		}
		logger.Log.Info("Using PostgreSQL storage")
	} else {
		URLRepository, err = repository.NewMemoryStorage(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Fatal("Error occurs while initializing file storage", zap.Error(err))
		}
		logger.Log.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
	}

	server := server.New(cfg.RunAddr)
	URLService := service.NewURLService(URLRepository, cfg.BaseURL)
	URLHandler := hdlr.NewURLHandler(URLService)

	server.HandleFunc("/", URLHandler.BaseHandler().ServeHTTP)
	server.HandleFunc("/{id}", URLHandler.BaseHandler().ServeHTTP)
	server.HandleFunc("/api/shorten", URLHandler.ShortenHandler().ServeHTTP)
	server.HandleFunc("/ping", URLHandler.PingHandler().ServeHTTP)
	server.HandleFunc("/api/shorten/batch", URLHandler.ShortenBatchHandler().ServeHTTP)
	server.HandleFunc("/api/getAll", URLHandler.GetAllHandler().ServeHTTP)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))

	err = server.Run()
	if err != nil {
		panic("Error occurs while running server: " + err.Error())
	}

}
