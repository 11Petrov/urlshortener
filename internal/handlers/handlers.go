package handlers

import (
	"io"
	"net/http"
	"strings"
)

// HandlerURLStore определяет приватный интерфейс для хранилища URL
type HandlerURLStore interface {
	ShortenURL(originalURL string) string
	RedirectURL(shortURL string) (string, error)
}

// URLHandler обрабатывает HTTP-запросы
type URLHandler struct {
	storeURL HandlerURLStore
	baseURL  string
}

// NewURLHandler создает новый экземпляр URLHandler
func NewURLHandler(storeURL HandlerURLStore, baseURL string) *URLHandler {
	return &URLHandler{
		storeURL: storeURL,
		baseURL:  baseURL,
	}
}

// ShortenURL обрабатывает запросы на сокращение URL
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
	shortURL := h.storeURL.ShortenURL(originalURL)
	responseURL := h.baseURL + "/" + shortURL

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(responseURL))
}

// RedirectURL обрабатывает запросы на перенаправление по сокращенному URL
func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	shortURL := path[1]
	if len(shortURL) == 0 {
		http.Error(w, "Empty URL parameter", http.StatusBadRequest)
		return
	}

	url, err := h.storeURL.RedirectURL(shortURL)
	if err != nil {
		http.Error(w, "Url not found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
