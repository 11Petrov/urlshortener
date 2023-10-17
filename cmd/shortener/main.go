package main

import (
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/gzip"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.NewConfig()
	log := logger.NewLogger()

	if err := Run(cfg, log); err != nil {
		log.Fatal(err)
	}
}

func Run(cfg *config.Config, log zap.SugaredLogger) error {
	storeURL := storage.NewRepo(cfg, log)
	h := handlers.NewHandlerURL(storeURL, cfg.BaseURL, log)
	r := chi.NewRouter()
	r.Use(logger.WithLogging)

	r.Post("/", gzip.GzipMiddleware(h.ShortenURL))
	r.Get("/{id}", gzip.GzipMiddleware(h.RedirectURL))
	r.Post("/api/shorten", gzip.GzipMiddleware(h.JSONShortenURL))
	r.Get("/ping", h.Ping)
	r.Post("/api/shorten/batch", gzip.GzipMiddleware(h.BatchShortenURL))
	log.Infow(
		"Running server",
		"address", cfg.ServerAddress,
		"DSN", cfg.DatabaseAddress,
	)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
