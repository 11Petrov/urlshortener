package handlers

import (
	"io"
	"net/http"

	"github.com/11Petrov/urlshortener/internal/storage"
)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ShortenURL(w, r)
	case http.MethodGet:
		RedirectURL(w, r)
	default:
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func ShortenURL(rw http.ResponseWriter, r *http.Request) {
	if r.ContentLength <= 0 {
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
	shortURL := GenerateShortURL(originalURL)
	storage.UrlMap[shortURL] = originalURL

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(storage.HostURL + shortURL))
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	if url, ok := storage.UrlMap[shortURL]; ok {
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Url not found", http.StatusBadRequest)
	}
}

func GenerateShortURL(url string) string {
	// Generating short URLs
	return "EwHXdJfB"
}
