package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

var AppConfig *Config

func NewConfig() *Config {
	cfg := &Config{}

	if cfg.ServerAddress = os.Getenv("SERVER_ADDRESS"); cfg.ServerAddress == "" {
		flag.StringVar(&cfg.ServerAddress, "a", ":8080", "адрес запуска HTTP-сервера")
		flag.Parse()
	}

	if cfg.BaseURL = os.Getenv("BASE_URL"); cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
		flag.Parse()
	}

	cfg.ServerAddress = strings.TrimPrefix(cfg.ServerAddress, "http://")

	AppConfig = cfg

	return cfg
}
