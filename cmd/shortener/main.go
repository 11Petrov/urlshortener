package main

import (
	"net/http"

	"github.com/11Petrov/urlshortener/config"
	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/go-chi/chi"
)

func main() {
	cfg := config.NewConfig()

	r := chi.NewRouter()

	r.Post("/", handlers.ShortenURL)
	r.Get("/{id}", handlers.RedirectURL)

	err := http.ListenAndServe(cfg.Addr, r)
	if err != nil {
		panic(err)
	}
}
