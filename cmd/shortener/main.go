package main

import (
	"log"
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/11Petrov/urlshortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.NewConfig()

	if err := Run(cfg); err != nil {
		panic(err)
	}
}

func Run(cfg *config.Config) error {
	storeURL := storage.NewRepoURL()
	h := handlers.NewHandlerURL(storeURL, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.RedirectURL)

	log.Println("Running server on", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
