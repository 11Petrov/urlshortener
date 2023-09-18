package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/11Petrov/urlshortener/internal/storage"
)

type URLHandlerInterface interface {
	ShortenURL(rw http.ResponseWriter, r *http.Request)
	RedirectURL(rw http.ResponseWriter, r *http.Request)
}

type URLHandler struct {
	dataStore storage.DataStore
	baseURL   string
}

func NewURLHandler(dataStore storage.DataStore, baseURL string) URLHandlerInterface {
	return &URLHandler{
		dataStore: dataStore,
		baseURL:   baseURL,
	}
}

// ShortenURL обрабатывает запросы на сокращение URL.
func (h *URLHandler) ShortenURL(rw http.ResponseWriter, r *http.Request) {
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
	shortURL := h.dataStore.ShortenURL(originalURL)
	responseURL := h.baseURL + "/" + shortURL

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(responseURL))
}

// RedirectURL обрабатывает запросы на перенаправление по сокращенному URL.
func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	shortURL := path[1]
	if len(shortURL) == 0 {
		http.Error(w, "Empty URL parameter", http.StatusBadRequest)
		return
	}

	url, err := h.dataStore.RedirectURL(shortURL)
	if err != nil {
		http.Error(w, "Url not found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
