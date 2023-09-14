package main

import (
	"log"
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/handlers"
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
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		handlers.ShortenURL(rw, r, cfg)
	})
	r.Get("/{id}", handlers.RedirectURL)

	log.Println("Running server on", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
