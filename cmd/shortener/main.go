package main

import (
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/gzip"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/storage"

	"github.com/go-chi/chi"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.NewConfig()

	if err := Run(cfg); err != nil {
		logger.Sugar.Fatal(err)
	}
}

func Run(cfg *config.Config) error {
	var h *handlers.HandlerURL
	if cfg.DatabaseAddress != "" {
		dbStoreURL, err := storage.NewDBStore(cfg.DatabaseAddress)
		if err != nil {
			logger.Sugar.Fatal(err)
		}
		h = handlers.NewHandlerURL(dbStoreURL, cfg.BaseURL)
	} else {
		storeURL, err := storage.NewRepoURL(cfg.FilePath)
		if err != nil {
			logger.Sugar.Fatal(err)
		}

		h = handlers.NewHandlerURL(storeURL, cfg.BaseURL)
	}
	r := chi.NewRouter()
	r.Use(logger.WithLogging)

	r.Post("/", gzip.GzipMiddleware(h.ShortenURL))
	r.Get("/{id}", gzip.GzipMiddleware(h.RedirectURL))
	r.Post("/api/shorten", gzip.GzipMiddleware(h.JSONShortenURL))
	r.Get("/ping", func(rw http.ResponseWriter, r *http.Request) {
		h.Ping(rw, r)
	})
	logger.Sugar.Infow(
		"Running server",
		"address", cfg.ServerAddress,
	)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
