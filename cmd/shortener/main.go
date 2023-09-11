package main

import (
	"log"
	"net/http"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/go-chi/chi"
)

func main() {
	cfg := config.NewConfig()
	log.Println(cfg.Addr, cfg.BaseURL)

	r := chi.NewRouter()

	r.Post("/", handlers.ShortenURL)
	r.Get("/{id}", handlers.RedirectURL)

	log.Println("Running server on", cfg.Addr)
	err := http.ListenAndServe(cfg.Addr, r)
	if err != nil {
		panic(err)
	}
}
