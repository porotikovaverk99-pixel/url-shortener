package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RunAddr         string        `json:"server_address" env:"SERVER_ADDRESS" env-default:":8080" flag:"a" flag-desc:"address and port to run server"`
	BaseURL         string        `json:"base_url" env:"BASE_URL" flag:"b" flag-desc:"base URL for shortened URLs"`
	LogLevel        string        `json:"log_level" env:"LOG_LEVEL" env-default:"Info" flag:"l" flag-desc:"log level"`
	FileStoragePath string        `json:"file_storage_path" env:"FILE_STORAGE_PATH" env-default:"storage.json" flag:"f" flag-desc:"file storage path"`
	DatabaseDSN     string        `json:"database_dsn" env:"DATABASE_DSN" flag:"d" flag-desc:"database dsn"`
	SecretKey       string        `json:"secret_key" env:"SECRET_KEY" flag:"k" flag-desc:"secret key"`
	DeleteQueueSize int           `json:"delete_queue_size" env:"DELETE_QUEUE_SIZE" env-default:"100" flag:"q" flag-desc:"delete queue size"`
	DeleteWorkers   int           `json:"delete_workers" env:"DELETE_WORKERS" env-default:"5" flag:"w" flag-desc:"delete workers count"`
	DeleteTimeout   time.Duration `json:"delete_timeout" env:"DELETE_TIMEOUT" env-default:"30s" flag:"t" flag-desc:"delete operation timeout"`
	FileAuditPath   string        `json:"audit_file" env:"AUDIT_FILE" env-default:"audit.json" flag:"audit-file" flag-desc:"file audit path"`
	URLAudit        string        `json:"audit_url" env:"AUDIT_URL" flag:"audit-url" flag-desc:"audit url"`
	EnableHTTPS     bool          `json:"enable_https" env:"ENABLE_HTTPS" env-default:"false" flag:"s" flag-desc:"enable HTTPS"`
	CertFile        string        `json:"cert_file" env:"TLS_CERT_FILE" env-default:"server.crt" flag:"cert" flag-desc:"TLS certificate file"`
	KeyFile         string        `json:"key_file" env:"TLS_KEY_FILE" env-default:"server.key" flag:"key" flag-desc:"TLS key file"`
	ConfigFile      string        `flag:"c" flag-desc:"path to config file"`
}

func ParseFlags() Config {
	var cfg Config

	configFileFlag := flag.String("c", "", "path to config file")
	flag.StringVar(configFileFlag, "config", "", "path to config file (alternative)")

	flag.StringVar(&cfg.RunAddr, "a", "", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "", "base URL for shortened URLs")
	flag.StringVar(&cfg.LogLevel, "l", "", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.SecretKey, "k", "", "secret key")
	flag.IntVar(&cfg.DeleteQueueSize, "q", 0, "delete queue size")
	flag.IntVar(&cfg.DeleteWorkers, "w", 0, "delete workers count")
	flag.DurationVar(&cfg.DeleteTimeout, "t", 0, "delete operation timeout")
	flag.StringVar(&cfg.FileAuditPath, "audit-file", "", "file audit path")
	flag.StringVar(&cfg.URLAudit, "audit-url", "", "audit url")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&cfg.CertFile, "cert", "", "TLS certificate file")
	flag.StringVar(&cfg.KeyFile, "key", "", "TLS key file")

	flag.Parse()

	configPath := *configFileFlag
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}
	cfg.ConfigFile = configPath

	if configPath != "" {
		file, err := os.Open(configPath)
		if err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&cfg); err != nil {
				log.Printf("Warning: failed to parse config file %s: %v", configPath, err)
			} else {
				log.Printf("Loaded config from %s", configPath)
			}
		} else {
			log.Printf("Warning: failed to open config file %s: %v", configPath, err)
		}
	}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	if cfg.RunAddr == "" {
		if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
			cfg.RunAddr = addr
		} else {
			cfg.RunAddr = ":8080"
		}
	}

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
