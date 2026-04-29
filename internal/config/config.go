package config

import (
	"flag"
	"strings"

	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RunAddr         string        `env:"SERVER_ADDRESS" env-default:":8080" flag:"a" flag-desc:"address and port to run server"`
	BaseURL         string        `env:"BASE_URL" flag:"b" flag-desc:"base URL for shortened URLs"`
	LogLevel        string        `env:"LOG_LEVEL" env-default:"Info" flag:"l" flag-desc:"log level"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH" env-default:"storage.json" flag:"f" flag-desc:"file storage path"`
	DatabaseDSN     string        `env:"DATABASE_DSN" flag:"d" flag-desc:"database dsn"`
	SecretKey       string        `env:"SECRET_KEY" flag:"k" flag-desc:"secret key"`
	DeleteQueueSize int           `env:"DELETE_QUEUE_SIZE" env-default:"100" flag:"q" flag-desc:"delete queue size"`
	DeleteWorkers   int           `env:"DELETE_WORKERS" env-default:"5" flag:"w" flag-desc:"delete workers count"`
	DeleteTimeout   time.Duration `env:"DELETE_TIMEOUT" env-default:"30s" flag:"t" flag-desc:"delete operation timeout"`
	FileAuditPath   string        `env:"AUDIT_FILE" env-default:"audit.json" flag:"audit-file" flag-desc:"file audit path"`
	URLAudit        string        `env:"AUDIT_URL" flag:"audit-url" flag-desc:"audit url"`
	EnableHTTPS     bool          `env:"ENABLE_HTTPS" env-default:"false" flag:"s" flag-desc:"enable HTTPS"`
	CertFile        string        `env:"TLS_CERT_FILE" env-default:"server.crt" flag:"cert" flag-desc:"TLS certificate file"`
	KeyFile         string        `env:"TLS_KEY_FILE" env-default:"server.key" flag:"key" flag-desc:"TLS key file"`
}

func ParseFlags() Config {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base URL for shortened URLs")
	flag.StringVar(&cfg.LogLevel, "l", cfg.LogLevel, "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")
	flag.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "secret key")
	flag.IntVar(&cfg.DeleteQueueSize, "q", cfg.DeleteQueueSize, "delete queue size")
	flag.IntVar(&cfg.DeleteWorkers, "w", cfg.DeleteWorkers, "delete workers count")
	flag.DurationVar(&cfg.DeleteTimeout, "t", cfg.DeleteTimeout, "delete operation timeout")
	flag.StringVar(&cfg.FileAuditPath, "audit-file", cfg.FileAuditPath, "file audit path")
	flag.StringVar(&cfg.URLAudit, "audit-url", cfg.URLAudit, "audit url")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "enable HTTPS")
	flag.StringVar(&cfg.CertFile, "cert", cfg.CertFile, "TLS certificate file")
	flag.StringVar(&cfg.KeyFile, "key", cfg.KeyFile, "TLS key file")

	flag.Parse()

	if cfg.BaseURL == "" {
		host := "localhost"
		if strings.HasPrefix(cfg.RunAddr, ":") {
			host += cfg.RunAddr
		} else {
			host = cfg.RunAddr
		}
		if cfg.EnableHTTPS {
			cfg.BaseURL = "https://" + host
		} else {
			cfg.BaseURL = "http://" + host
		}
	}

	if !strings.HasPrefix(cfg.BaseURL, "http://") && !strings.HasPrefix(cfg.BaseURL, "https://") {
		if cfg.EnableHTTPS {
			cfg.BaseURL = "https://" + cfg.BaseURL
		} else {
			cfg.BaseURL = "http://" + cfg.BaseURL
		}
	}

	return cfg
}
