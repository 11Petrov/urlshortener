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
	log.Println(cfg.ServerAddress, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", handlers.ShortenURL)
	r.Get("/{id}", handlers.RedirectURL)

	log.Println("Running server on", cfg.ServerAddress)
	err := http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
