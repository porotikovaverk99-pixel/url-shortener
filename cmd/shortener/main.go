package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	URLService := service.NewURLService(URLRepository, cfg.BaseURL, 100, 5)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		logger.Log.Info("Shutdown signal received")
		URLService.Shutdown()
		os.Exit(0)
	}()

	URLHandler := hdlr.NewURLHandler(URLService)

	server := svr.New(cfg.RunAddr)

	router := server.Router()

	secretKey := cfg.SecretKey
	router.Use(auth.Auth(secretKey))
	router.Use(logger.RequestLogger)
	router.Use(packgzip.GzipMiddleware)

	server.HandleFunc("/", URLHandler.BaseHandler().ServeHTTP)
	server.HandleFunc("/{id}", URLHandler.BaseHandler().ServeHTTP)
	server.HandleFunc("/api/shorten", URLHandler.ShortenHandler().ServeHTTP)
	server.HandleFunc("/ping", URLHandler.PingHandler().ServeHTTP)
	server.HandleFunc("/api/shorten/batch", URLHandler.ShortenBatchHandler().ServeHTTP)
	server.HandleFunc("/api/user/urls", URLHandler.UserUrlsHandler().ServeHTTP)

	logger.Log.Info("Running server", zap.String("address", cfg.RunAddr))

	err = server.Run()
	if err != nil {
		logger.Log.Fatal("Error occurs while running server", zap.Error(err))
	}

}
