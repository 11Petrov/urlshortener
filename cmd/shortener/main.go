package main

import (
	"net/http"

	"github.com/11Petrov/urlshortener/internal/handlers"
)

func main() {
	http.HandleFunc("/", handlers.HandleRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
