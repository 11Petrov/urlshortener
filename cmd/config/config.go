package config

import (
	"flag"
	"os"
	"strings"
)

// Config содержит конфигурационные параметры приложения.
type Config struct {
	ServerAddress string
	BaseURL       string
}

// NewConfig создает новый экземпляр конфигурации приложения на основе флагов командной строки и переменных окружения.
func NewConfig() *Config {
	serverAddressFlag := flag.String("a", "localhost:8080", "адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
	flag.Parse()

	cfg := &Config{}

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	} else {
		cfg.ServerAddress = *serverAddressFlag
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	} else {
		cfg.BaseURL = *baseURLFlag
	}

	cfg.ServerAddress = strings.TrimPrefix(cfg.ServerAddress, "http://")
	parts := strings.Split(cfg.ServerAddress, ":")
	if parts[0] == "" {
		cfg.ServerAddress = "localhost:" + parts[1]
	}

	return cfg
}
