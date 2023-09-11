package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/storage"
	"github.com/go-chi/chi"
)

// ShortenURL обрабатывает запросы на сокращение URL.
func ShortenURL(rw http.ResponseWriter, r *http.Request) {
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
	log.Println("OriginalURL: ", originalURL)
	shortURL := GenerateShortURL(originalURL)
	log.Println("ShortURL: ", shortURL)
	storage.URLMap[shortURL] = originalURL

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte("http://" + config.AppConfig.ServerAddress + "/" + shortURL))
}

// RedirectURL обрабатывает запросы на перенаправление по сокращенному URL.
func RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if len(shortURL) == 0 {
		http.Error(w, "Empty URL parametr", http.StatusBadRequest)
		return
	}

	if url, ok := storage.URLMap[shortURL]; ok {
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)

		log.Println("StatusCode:", http.StatusTemporaryRedirect)
		log.Println("Location:", url)
	} else {
		http.Error(w, "Url not found", http.StatusBadRequest)
	}
}

// GenerateShortURL генерирует сокращенный URL на основе хэша.
func GenerateShortURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	shortURL := base64.URLEncoding.EncodeToString(hash[:])
	regExp := regexp.MustCompile("[^a-zA-Z0-9]+")
	shortURL = regExp.ReplaceAllString(shortURL, "")
	return shortURL[:8]
}
