package config

import (
	"flag"
	"strings"

	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RunAddr         string `env:"SERVER_ADDRESS" env-default:":8080" flag:"a" flag-desc:"address and port to run server"`
	BaseURL         string `env:"BASE_URL" flag:"b" flag-desc:"base URL for shortened URLs"`
	LogLevel        string `env:"LOG_LEVEL" env-default:"Info" flag:"l" flag-desc:"log level"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" env-default:"storage.json" flag:"f" flag-desc:"file storage path"`
	DatabaseDSN     string `env:"DATABASE_DSN" flag:"d" flag-desc:"database dsn"`
	SecretKey       string `env:"SECRET_KEY" flag:"k" flag-desc:"secret key"`
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

	flag.Parse()

	if cfg.BaseURL == "" {
		host := "localhost"
		if strings.HasPrefix(cfg.RunAddr, ":") {
			host += cfg.RunAddr
		} else {
			host = cfg.RunAddr
		}
		cfg.BaseURL = "http://" + host
	}

	if !strings.HasPrefix(cfg.BaseURL, "http://") && !strings.HasPrefix(cfg.BaseURL, "https://") {
		cfg.BaseURL = "http://" + cfg.BaseURL
	}

	return cfg
}
