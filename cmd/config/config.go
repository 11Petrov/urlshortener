package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

var AppConfig *Config

func NewConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	if cfg.ServerAddress = os.Getenv("SERVER_ADDRESS"); cfg.ServerAddress == "" {
		flag.StringVar(&cfg.ServerAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	}

	if cfg.BaseURL = os.Getenv("BASE_URL"); cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
	}

	AppConfig = cfg

	return cfg
}
