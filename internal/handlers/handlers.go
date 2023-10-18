package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/models"
	e "github.com/11Petrov/urlshortener/internal/storage/errors"
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
	log := logger.LoggerFromContext(r.Context())
	if r.ContentLength == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Error("Request body is missing (ShortenURL)")
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Error reading request (ShortenURL) %s", err)
		return
	}

	originalURL := string(body)
	shortURL, err := h.storeURL.ShortenURL(r.Context(), originalURL)
	if err != nil {
		if err == e.ErrUnique {
			rw.WriteHeader(http.StatusConflict)
			responseURL := h.baseURL + "/" + shortURL
			rw.Write([]byte(responseURL))
			log.Errorf("URL already in database (ShortenURL) %s", err)
			return
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			log.Errorf("ShortenURL error %s", err)
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
	log := logger.LoggerFromContext(r.Context())
	log.Info("Processing RediretURL handler")
	path := strings.Split(r.URL.Path, "/")
	shortURL := path[1]
	if len(shortURL) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Error("Empty URL parameter (RedirectURL)")
		return
	}

	url, err := h.storeURL.RedirectURL(r.Context(), shortURL)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Url not found (RedirectURL) %s", err)
		return
	}
	rw.Header().Set("Location", url)
	rw.WriteHeader(http.StatusTemporaryRedirect)
}

// JSONShortenURL обрабатывает запросы на сокращение URL и возвращает JSON-ответ
func (h *HandlerURL) JSONShortenURL(rw http.ResponseWriter, r *http.Request) {
	log := logger.LoggerFromContext(r.Context())
	var req models.JSONShortenURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Invalid decode json (JSONShortenURL) %s", err)
		return
	}
	shortURL, err := h.storeURL.ShortenURL(r.Context(), req.URL)
	if err != nil {
		if err == e.ErrUnique {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusConflict)
			resp := models.JSONShortenURLResponse{Result: h.baseURL + "/" + shortURL}
			log.Errorf("URL already in database (JSONShortenURL) %s", err)
			if err := json.NewEncoder(rw).Encode(resp); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				log.Errorf("Invalid encode json (JSONShortenURL) %s", err)
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
			rw.WriteHeader(http.StatusBadRequest)
			log.Errorf("Invalid encode json (JSONShortenURL) %s", err)
			return
		}
	}
}

func (h *HandlerURL) Ping(rw http.ResponseWriter, r *http.Request) {
	log := logger.LoggerFromContext(r.Context())
	err := h.storeURL.Ping(r.Context())
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Database connection failed (Ping) %s", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *HandlerURL) BatchShortenURL(rw http.ResponseWriter, r *http.Request) {
	log := logger.LoggerFromContext(r.Context())
	var arrRequest []models.BatchRequest
	var arrResponse []models.BatchResponse

	if err := json.NewDecoder(r.Body).Decode(&arrRequest); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Invalid decode json (BatchShortenURL) %s", err)
		return
	}

	for _, val := range arrRequest {
		shortURL, err := h.storeURL.BatchShortenURL(r.Context(), val.OriginalURL)
		if err != nil {
			log.Errorf("BatchShortenURL error %s", err)
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
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Invalid encode json (BatchShortenURL) %s", err)
		return
	}
}
