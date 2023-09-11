package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var AppConfig *Config

func NewConfig() *Config {
	cfg := &Config{}

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		cfg.ServerAddress = serverAddress
	} else {
		flag.StringVar(&cfg.ServerAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	} else {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
	}

	flag.Parse()

	AppConfig = cfg

	return cfg
}
