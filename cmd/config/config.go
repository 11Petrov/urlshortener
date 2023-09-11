package config

import "flag"

type Config struct {
	Addr    string
	BaseURL string
}

var AppConfig *Config

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")

	flag.Parse()

	AppConfig = cfg

	return cfg
}
