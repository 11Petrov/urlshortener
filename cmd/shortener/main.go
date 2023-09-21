package main

import (
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/storage"
	"github.com/go-chi/chi"
)

func main() {
	cfg := config.NewConfig()

	if err := Run(cfg); err != nil {
		logger.Sugar.Fatal(err)
		panic(err)
	}
}

func Run(cfg *config.Config) error {
	storeURL := storage.NewRepoURL()
	h := handlers.NewHandlerURL(storeURL, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(logger.WithLogging)

	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.RedirectURL)

	logger.Sugar.Infow(
		"Running server",
		"address", cfg.ServerAddress,
	)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
