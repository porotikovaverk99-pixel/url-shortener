package main

import (
	"log"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/auth"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
	packgzip "github.com/porotikovaverk99-pixel/url-shortener/internal/gzip"
	hdlr "github.com/porotikovaverk99-pixel/url-shortener/internal/handler"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/logger"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	svr "github.com/porotikovaverk99-pixel/url-shortener/internal/server"
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
	var UserRepository repository.UserRepository
	var err error

	if cfg.DatabaseDSN != "" {
		storage, err := repository.NewPostgresStorage(cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Error occurs while initializing PostgreSQL storage", zap.Error(err))
		}
		logger.Log.Info("Using PostgreSQL storage")
		URLRepository = storage
		UserRepository = storage
	} else {
		storage, err := repository.NewMemoryStorage(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Fatal("Error occurs while initializing file storage", zap.Error(err))
		}
		logger.Log.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
		URLRepository = storage
		UserRepository = storage
	}

	URLService := service.NewURLService(URLRepository, cfg.BaseURL)
	URLHandler := hdlr.NewURLHandler(URLService)

	server := svr.New(cfg.RunAddr)

	router := server.Router()

	secretKey := cfg.SecretKey
	router.Use(auth.Auth(secretKey, UserRepository))
	router.Use(logger.RequestLogger)
	router.Use(packgzip.GzipMiddleware)

	server.HandleFunc("/", URLHandler.BaseHandler().ServeHTTP)
	server.HandleFunc("/{id}", URLHandler.BaseHandler().ServeHTTP)
	server.HandleFunc("/api/shorten", URLHandler.ShortenHandler().ServeHTTP)
	server.HandleFunc("/ping", URLHandler.PingHandler().ServeHTTP)
	server.HandleFunc("/api/shorten/batch", URLHandler.ShortenBatchHandler().ServeHTTP)
	server.HandleFunc("/api/user/urls", URLHandler.GetAllHandler().ServeHTTP)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))

	err = server.Run()
	if err != nil {
		logger.Log.Fatal("Error occurs while running server", zap.Error(err))
	}

}
