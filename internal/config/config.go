package config

import (
	"flag"
	"strings"
)

type Config struct {
    RunAddr string
    BaseURL string
}

func ParseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "", "base URL for shortened URLs")
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