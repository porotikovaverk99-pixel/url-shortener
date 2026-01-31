package main

import (
	"log"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
	packgzip "github.com/porotikovaverk99-pixel/url-shortener/internal/gzip"
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
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Log.Sync()

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

	baseHandler := logger.RequestLogger(packgzip.GzipMiddleware(URLHandler.BaseHandler()))
	shortenHandler := logger.RequestLogger(packgzip.GzipMiddleware(URLHandler.ShortenHandler()))
	pingHandler := logger.RequestLogger(packgzip.GzipMiddleware(URLHandler.PingHandler()))
	batchHandler := logger.RequestLogger(packgzip.GzipMiddleware(URLHandler.ShortenBatchHandler()))
	getAllHandler := logger.RequestLogger(packgzip.GzipMiddleware(URLHandler.GetAllHandler()))

	server.HandleFunc("/", baseHandler.ServeHTTP)
	server.HandleFunc("/{id}", baseHandler.ServeHTTP)
	server.HandleFunc("/api/shorten", shortenHandler.ServeHTTP)
	server.HandleFunc("/ping", pingHandler.ServeHTTP)
	server.HandleFunc("/api/shorten/batch", batchHandler.ServeHTTP)
	server.HandleFunc("/api/user/urls", getAllHandler.ServeHTTP)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))

	err = server.Run()
	if err != nil {
		logger.Log.Fatal("Error occurs while running server", zap.Error(err))
	}

}
