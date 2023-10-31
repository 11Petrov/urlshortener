package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/11Petrov/urlshortener/internal/auth"
	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/models"
	storageErrors "github.com/11Petrov/urlshortener/internal/storage/errors"
)

// handlerURLStore определяет приватный интерфейс для хранилища URL
type handlerURLStore interface {
	ShortenURL(ctx context.Context, userID, originalURL string) (string, error)
	RedirectURL(ctx context.Context, userID, shortURL string) (string, error)
	Ping(ctx context.Context) error
	BatchShortenURL(ctx context.Context, userID, originalURL string) (string, error)
	GetUserURLs(ctx context.Context, userID, baseURL string) ([]models.Event, error)
	DeleteUserURLs(ctx context.Context, userID string, shortURL []string) error
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

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		log.Error("error userID ShortenURL")
	}
	originalURL := string(body)
	shortURL, err := h.storeURL.ShortenURL(r.Context(), userID, originalURL)
	if err != nil {
		if err == storageErrors.ErrUnique {
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
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*60)
	defer cancel()
	path := strings.Split(r.URL.Path, "/")
	shortURL := path[1]
	if len(shortURL) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Error("Empty URL parameter (RedirectURL)")
		return
	}

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		log.Error("error userID RedirectURL")
	}
	url, err := h.storeURL.RedirectURL(ctx, userID, shortURL)
	if err != nil {
		log.Errorf("URL not found (RedirectURL) %s", err)
		rw.WriteHeader(http.StatusGone)
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
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		log.Error("error userID JsonShortenURL")
		return
	}
	shortURL, err := h.storeURL.ShortenURL(r.Context(), userID, req.URL)
	if err != nil {
		if err == storageErrors.ErrUnique {
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
		userID, ok := r.Context().Value(auth.UserIDKey).(string)
		if !ok {
			log.Error("error userID BatchShortenURL")
		}
		shortURL, err := h.storeURL.BatchShortenURL(r.Context(), userID, val.OriginalURL)
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

func (h *HandlerURL) GetUserURLs(rw http.ResponseWriter, r *http.Request) {
	log := logger.LoggerFromContext(r.Context())

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.storeURL.GetUserURLs(r.Context(), userID, h.baseURL)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Errorf("GetUserURLs error %s", err)
		return
	}

	if len(urls) == 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(rw).Encode(urls); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Invalid encode json (GetUserUrls) %s", err)
		return
	}
}

func (h *HandlerURL) DeleteUserURLs(rw http.ResponseWriter, r *http.Request) {
	log := logger.LoggerFromContext(r.Context())
	log.Info("Processing DeleteUserURLs handelr")
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*60)
	defer cancel()

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		log.Error("Error getting user ID in DeleteUserURLs")
		return
	}

	var urls []string
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Errorf("Invalid decode json (BatchShortenURL) %s", err)
		return
	}

	err := h.storeURL.DeleteUserURLs(ctx, userID, urls)
	if err != nil {
		log.Errorf("DeleteUserURLs error", err)
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusAccepted)
}
