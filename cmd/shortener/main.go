package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/audit"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/config"
	packgzip "github.com/porotikovaverk99-pixel/url-shortener/internal/gzip"
	hdlr "github.com/porotikovaverk99-pixel/url-shortener/internal/handler"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/logger"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/middleware"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	svr "github.com/porotikovaverk99-pixel/url-shortener/internal/server"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/service"
	"go.uber.org/zap"
)

const (
	shutdownTimeout       = 15 * time.Second
	serverShutdownTimeout = 10 * time.Second
)

func main() {

	cfg := config.ParseFlags()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Log.Sync()
	}()

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

	auditManager := audit.NewManager()

	if cfg.FileAuditPath != "" {
		auditManager.AddObserver(audit.NewFileObserver(cfg.FileAuditPath))
		log.Printf("File audit enabled: %s", cfg.FileAuditPath)
	}

	if cfg.URLAudit != "" {
		auditManager.AddObserver(audit.NewHTTPObserver(cfg.URLAudit))
		log.Printf("HTTP audit enabled: %s", cfg.URLAudit)
	}

	URLService := service.NewURLService(
		URLRepository,
		cfg.BaseURL,
		cfg.DeleteQueueSize,
		cfg.DeleteWorkers,
		cfg.DeleteTimeout,
		logger.Log,
	)

	URLHandler := hdlr.NewURLHandler(URLService)
	server := svr.New(cfg.RunAddr)

	router := server.Router()
	secretKey := cfg.SecretKey

	router.Use(middleware.Auth(secretKey))
	router.Use(logger.RequestLogger)
	router.Use(packgzip.GzipMiddleware)

	auditMid := middleware.AuditMiddleware(auditManager)

	server.HandleFunc("/", auditMid(URLHandler.BaseHandler()).ServeHTTP)
	server.HandleFunc("/{id}", auditMid(URLHandler.BaseHandler()).ServeHTTP)
	server.HandleFunc("/api/shorten", auditMid(URLHandler.ShortenHandler()).ServeHTTP)

	server.HandleFunc("/ping", URLHandler.PingHandler().ServeHTTP)
	server.HandleFunc("/api/shorten/batch", URLHandler.ShortenBatchHandler().ServeHTTP)
	server.HandleFunc("/api/user/urls", URLHandler.UserUrlsHandler().ServeHTTP)

	serverErr := make(chan error, 1)
	go func() {
		logger.Log.Info("Starting server", zap.String("address", cfg.RunAddr))
		if err := server.Run(); err != nil {
			serverErr <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Log.Error("Server stopped with error", zap.Error(err))
		gracefulShutdown(server, URLService, URLRepository, logger.Log)
		os.Exit(1)

	case sig := <-sigChan:
		logger.Log.Info("Received shutdown signal", zap.String("signal", sig.String()))
		gracefulShutdown(server, URLService, URLRepository, logger.Log)
		logger.Log.Info("Application shutdown completed")
	}

}

func gracefulShutdown(
	server *svr.Server,
	service *service.URLService,
	repo repository.URLRepository,
	log *zap.Logger,
) {
	log.Info("Starting graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Info("Shutting down HTTP server...")
	serverCtx, serverCancel := context.WithTimeout(shutdownCtx, serverShutdownTimeout)
	defer serverCancel()

	if err := server.Shutdown(serverCtx); err != nil {
		log.Error("HTTP server shutdown error", zap.Error(err))
	} else {
		log.Info("HTTP server stopped gracefully")
	}

	log.Info("Shutting down background workers...")
	service.Shutdown()
	log.Info("Background workers stopped")

	log.Info("Closing repository...")
	if err := repo.Close(); err != nil {
		log.Error("Failed to close repository", zap.Error(err))
	} else {
		log.Info("Repository closed")
	}

	select {
	case <-shutdownCtx.Done():
		log.Warn("Shutdown timed out - some operations may not have completed")
	default:
		log.Info("Graceful shutdown completed successfully")
	}
}
