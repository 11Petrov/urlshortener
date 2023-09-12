package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.NewConfig()
	flag.Parse()
	config.Set(cfg)

	if err := Run(cfg); err != nil {
		panic(err)
	}
}

func Run(cfg *config.Config) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", handlers.ShortenURL)
	r.Get("/{id}", handlers.RedirectURL)

	log.Println("Running server on", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
