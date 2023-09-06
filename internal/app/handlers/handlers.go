package handlers

import (
	"fmt"
	"io"
	"net/http"
)

const (
	hostURL = "http://localhost:8080/"
)

var urlMap = make(map[string]string)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		shortenURL(w, r)
	case http.MethodGet:
		redirectURL(w, r)
	default:
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()
	originalURL := string(body)
	shortURL := generateShortURL(originalURL)
	urlMap[shortURL] = originalURL

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(hostURL + shortURL))
}

func redirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	if url, ok := urlMap[shortURL]; ok {
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Url not found", http.StatusBadRequest)
	}
}

func generateShortURL(url string) string {
	// Логика генерации коротких URL
	return "EwHXdJfB"
}
