package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/11Petrov/urlshortener/cmd/config"
	storage "github.com/11Petrov/urlshortener/internal/storage/urls"
)

var urlStorage = storage.NewStorageURLMap()

// ShortenURL обрабатывает запросы на сокращение URL.
func ShortenURL(rw http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.ContentLength == 0 {
		http.Error(rw, "Request body is missing", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "Error reading request", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	shortURL := urlStorage.SetURL(originalURL)
	responseURL := "http://" + cfg.ServerAddress + "/" + shortURL

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(responseURL))
}

// RedirectURL обрабатывает запросы на перенаправление по сокращенному URL.
func RedirectURL(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	shortURL := path[1]
	if len(shortURL) == 0 {
		http.Error(w, "Empty URL parameter", http.StatusBadRequest)
		return
	}

	url, err := urlStorage.GetURL(shortURL)
	if err != nil {
		http.Error(w, "Url not found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
