package main

import (
	"context"
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/auth"
	"github.com/11Petrov/urlshortener/internal/gzip"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/11Petrov/urlshortener/internal/logger"
	_ "github.com/11Petrov/urlshortener/internal/migrations"
	"github.com/11Petrov/urlshortener/internal/storage"

	"github.com/go-chi/chi"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.NewConfig()
	log := logger.NewLogger()

	ctx := logger.ContextWithLogger(context.Background(), &log)

	if err := Run(cfg, ctx); err != nil {
		log.Fatal(err)
	}
}

func Run(cfg *config.Config, ctx context.Context) error {
	log := logger.LoggerFromContext(ctx)
	storeURL := storage.NewRepo(cfg, ctx)
	h := handlers.NewHandlerURL(storeURL, cfg.BaseURL)
	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Use(auth.AuthMiddleware)

	r.Post("/", gzip.GzipMiddleware(h.ShortenURL))
	r.Get("/{id}", gzip.GzipMiddleware(h.RedirectURL))
	r.Post("/api/shorten", gzip.GzipMiddleware(h.JSONShortenURL))
	r.Get("/ping", h.Ping)
	r.Post("/api/shorten/batch", gzip.GzipMiddleware(h.BatchShortenURL))
	r.Get("/api/user/urls", gzip.GzipMiddleware(h.GetUserURLs))
	log.Infow(
		"Running server",
		"address", cfg.ServerAddress,
		"DSN", cfg.DatabaseAddress,
	)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
