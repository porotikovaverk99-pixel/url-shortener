package config

import "flag"

type Config struct {
    RunAddr string
    BaseURL string
}

func ParseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	flag.Parse()

	return cfg
} 