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

var (
	serverAddresFlag string
	baseURLFlag      string
)

func init() {
	flag.StringVar(&serverAddresFlag, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&baseURLFlag, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
}

var AppConfig *Config

func NewConfig() *Config {
	flag.Parse()

	cfg := &Config{}

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	} else {
		cfg.ServerAddress = serverAddresFlag
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	} else {
		cfg.BaseURL = baseURLFlag
	}

	AppConfig = cfg

	return cfg
}

func Set(c *Config) string {
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")

	parts := strings.Split(c.ServerAddress, ":")
	if parts[0] == "" {
		c.ServerAddress = "localhost:" + parts[1]
	}
	log.Println(c.ServerAddress, c.BaseURL)
	return c.ServerAddress
}
