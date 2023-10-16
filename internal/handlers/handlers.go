package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/11Petrov/urlshortener/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// handlerURLStore определяет приватный интерфейс для хранилища URL
type handlerURLStore interface {
	ShortenURL(ctx context.Context, originalURL string) (string, error)
	RedirectURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	BatchShortenURL(ctx context.Context, originalURL string) (string, error)
}

// URLHandler обрабатывает HTTP-запросы
type HandlerURL struct {
	storeURL handlerURLStore
	baseURL  string
}

// NewURLHandler создает новый экземпляр URLHandler
func NewHandlerURL(storeURL handlerURLStore, baseURL string) *HandlerURL {
	return &HandlerURL{
		storeURL: storeURL,
		baseURL:  baseURL,
	}
}

// ShortenURL обрабатывает запросы на сокращение URL
func (h *HandlerURL) ShortenURL(rw http.ResponseWriter, r *http.Request) {
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
	shortURL, err := h.storeURL.ShortenURL(r.Context(), originalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			rw.WriteHeader(http.StatusConflict)
			responseURL := h.baseURL + "/" + shortURL
			rw.Write([]byte(responseURL))
			return
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	responseURL := h.baseURL + "/" + shortURL

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(responseURL))
}

// RedirectURL обрабатывает запросы на перенаправление по сокращенному URL
func (h *HandlerURL) RedirectURL(rw http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	shortURL := path[1]
	if len(shortURL) == 0 {
		http.Error(rw, "Empty URL parameter", http.StatusBadRequest)
		return
	}

	url, err := h.storeURL.RedirectURL(r.Context(), shortURL)
	if err != nil {
		http.Error(rw, "Url not found", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Location", url)
	rw.WriteHeader(http.StatusTemporaryRedirect)
}

// JSONShortenURL обрабатывает запросы на сокращение URL и возвращает JSON-ответ
func (h *HandlerURL) JSONShortenURL(rw http.ResponseWriter, r *http.Request) {
	var req models.JSONShortenURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "Invalid decode json", http.StatusBadRequest)
		return
	}
	shortURL, err := h.storeURL.ShortenURL(r.Context(), req.URL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusConflict)
			resp := models.JSONShortenURLResponse{Result: h.baseURL + "/" + shortURL}
			if err := json.NewEncoder(rw).Encode(resp); err != nil {
				http.Error(rw, "Invalid encode json", http.StatusBadRequest)
				return
			}
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		resp := models.JSONShortenURLResponse{Result: h.baseURL + "/" + shortURL}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			http.Error(rw, "Invalid encode json", http.StatusBadRequest)
			return
		}
	}
}

func (h *HandlerURL) Ping(rw http.ResponseWriter, r *http.Request) {
	err := h.storeURL.Ping(r.Context())
	if err != nil {
		http.Error(rw, "Database connection failed", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *HandlerURL) BatchShortenURL(rw http.ResponseWriter, r *http.Request) {
	var arrRequest []models.BatchRequest
	var arrResponse []models.BatchResponse

	if err := json.NewDecoder(r.Body).Decode(&arrRequest); err != nil {
		http.Error(rw, "Invalid decode json", http.StatusBadRequest)
		return
	}

	for _, val := range arrRequest {
		shortURL, err := h.storeURL.BatchShortenURL(r.Context(), val.OriginalURL)
		if err != nil {
			return
		}
		url := h.baseURL + "/" + shortURL
		resp := models.BatchResponse{
			CorrelationID: val.CorrelationID,
			ShortURL:      url,
		}

		arrResponse = append(arrResponse, resp)
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(rw).Encode(&arrResponse); err != nil {
		http.Error(rw, "Invalid encode json", http.StatusBadRequest)
		return
	}
}
