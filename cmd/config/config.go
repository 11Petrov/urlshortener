package config

import (
	"flag"
	"log"
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
		flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	}

	if cfg.BaseURL = os.Getenv("BASE_URL"); cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
	}

	flag.Parse()

	AppConfig = cfg

	return cfg
}

func Set(c *Config) string {
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")

	parts := strings.Split(c.ServerAddress, ":")
	if parts[0] == "" {
		c.ServerAddress = "localhost:" + parts[1]
	}
	log.Println(parts)
	log.Println(c.ServerAddress)
	return c.ServerAddress
}
