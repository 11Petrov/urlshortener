package config

import (
	"flag"
	"log"
	"os"
	"strings"
)

// Config содержит конфигурационные параметры приложения.
type Config struct {
	ServerAddress string
	BaseURL       string
}

var (
	serverAddresFlag string
	baseURLFlag      string
)

// Инициализация флагов командной строки.
func init() {
	flag.StringVar(&serverAddresFlag, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&baseURLFlag, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
}

var AppConfig *Config

// NewConfig создает новый экземпляр конфигурации приложения на основе флагов командной строки и переменных окружения.
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

// Set обновляет конфигурацию и возвращает отформатированный адрес сервера.
func Set(c *Config) string {
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")

	parts := strings.Split(c.ServerAddress, ":")
	if parts[0] == "" {
		c.ServerAddress = "localhost:" + parts[1]
	}
	log.Println(c.ServerAddress, c.BaseURL)
	return c.ServerAddress
}
