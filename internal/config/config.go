package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	RunAddr         string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DatabaseDSN     string
}

func ParseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "", "base URL for shortened URLs")
	flag.StringVar(&cfg.LogLevel, "l", "Info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "postgres://postgres:123@localhost:5432/url_shortener?sslmode=disable", "database dsn")
	flag.Parse()

	if sa := os.Getenv("SERVER_ADDRESS"); sa != "" {
		cfg.RunAddr = sa
	}

	if bu := os.Getenv("BASE_URL"); bu != "" {
		cfg.BaseURL = bu
	}

	if l := os.Getenv("LOG_LEVEL"); l != "" {
		cfg.LogLevel = l
	}

	if f := os.Getenv("FILE_STORAGE_PATH"); f != "" {
		cfg.FileStoragePath = f
	}

	if dd := os.Getenv("DATABASE_DSN"); dd != "" {
		cfg.DatabaseDSN = dd
	}

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
