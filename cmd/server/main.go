package main

import (
	"net/http"

	"github.com/11Petrov/urlshortener/internal/handlers"
	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	r.Post("/", handlers.ShortenURL)
	r.Get("/{id}", handlers.RedirectURL)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
