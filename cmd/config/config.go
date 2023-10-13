package config

import (
	"flag"
	"os"
	"strings"
)

// Config содержит конфигурационные параметры приложения
type Config struct {
	ServerAddress   string
	BaseURL         string
	FilePath        string
	DatabaseAddress string
}

// parseFlags обрабатывает флаги командной строки и возвращает значения по умолчанию, если флаги не установлены
func parseFlags() (string, string, string, string) {
	serverAddressFlag := flag.String("a", "localhost:8080", "адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "http://localhost:8080", "базовый адрес результирующего сокращённого URL")
	filePathFlag := flag.String("f", "/tmp/short-url-db.json", "полное имя файла для сохранения данных в формате JSON")
	databaseAddressFlag := flag.String("d", "", "Database address")

	flag.Parse()
	return *serverAddressFlag, *baseURLFlag, *filePathFlag, *databaseAddressFlag
}

// parseEnv обрабатывает переменные окружения и возвращает их значения
func parseEnv() (string, string, string, string) {
	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envFilePath := os.Getenv("FILE_STORAGE_PATH")
	envDatabaseAddress := os.Getenv("DATABASE_DSN")
	return envServerAddress, envBaseURL, envFilePath, envDatabaseAddress
}

// NewConfig создает новый экземпляр конфигурации приложения на основе флагов командной строки и переменных окружения
func NewConfig() *Config {
	serverAddressFlag, baseURLFlag, filePathFlag, databaseAddressFlag := parseFlags()
	envServerAddress, envBaseURL, envFilePath, envDatabaseAddress := parseEnv()

	cfg := &Config{}

	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	} else {
		cfg.ServerAddress = serverAddressFlag
	}

	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	} else {
		cfg.BaseURL = baseURLFlag
	}

	if envFilePath != "" {
		cfg.FilePath = envFilePath
	} else {
		cfg.FilePath = filePathFlag
	}

	if envDatabaseAddress != "" {
		cfg.DatabaseAddress = envDatabaseAddress
	} else {
		cfg.DatabaseAddress = databaseAddressFlag
	}

	cfg.ServerAddress = strings.TrimPrefix(cfg.ServerAddress, "http://")
	parts := strings.Split(cfg.ServerAddress, ":")
	if parts[0] == "" {
		cfg.ServerAddress = "localhost:" + parts[1]
	}

	return cfg
}
