package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/auth"
	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/models"
	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestURLStore interface {
	ShortenURL(ctx context.Context, userID, originalURL string) (string, error)
	RedirectURL(ctx context.Context, userID, shortURL string) (string, error)
	Ping(ctx context.Context) error
	BatchShortenURL(ctx context.Context, userID, originalURL string) (string, error)
	GetUserURLs(ctx context.Context, userID, baseURL string) ([]models.Event, error)
	DeleteUserURLs(ctx context.Context, userID string, shortURL []string) error
}

type testStorage struct {
	URLMap map[string]map[string]string
}

func newTestStorage() TestURLStore {
	return &testStorage{
		URLMap: make(map[string]map[string]string),
	}
}

func (t *testStorage) ShortenURL(ctx context.Context, userID, originalURL string) (string, error) {
	shortURL := utils.GenerateShortURL(originalURL)
	if _, ok := t.URLMap[userID]; !ok {
		t.URLMap[userID] = make(map[string]string)
	}
	t.URLMap[userID][shortURL] = originalURL
	return shortURL, nil
}

func (t *testStorage) RedirectURL(ctx context.Context, userID, shortURL string) (string, error) {
	if userUrls, ok := t.URLMap[userID]; ok {
		url, ok := userUrls[shortURL]
		if !ok {
			return "", errors.New("url not found for this user")
		}
		return url, nil
	}
	return "", errors.New("user not found")
}

func (t *testStorage) GetUserURLs(ctx context.Context, userID, baseURL string) ([]models.Event, error) {
	log := logger.LoggerFromContext(ctx)

	urls, ok := t.URLMap[userID]
	if !ok {
		log.Info("No URLs found for the user")
		return []models.Event{}, nil
	}

	events := make([]models.Event, 0, len(urls))

	for shortURL, originalURL := range urls {
		events = append(events, models.Event{
			UserID:      userID,
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}

	return events, nil
}

// BatchShortenURL implements URLStore.
func (t *testStorage) BatchShortenURL(ctx context.Context, userID, originalURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	log.Info("BatchShortenURL function was called")
	return "", nil
}

// Ping implements URLStore.
func (t *testStorage) Ping(ctx context.Context) error {
	log := logger.LoggerFromContext(ctx)
	log.Info("Ping function was called")
	return nil
}

func (t *testStorage) DeleteUserURLs(ctx context.Context, userID string, shortURL []string) error {
	log := logger.LoggerFromContext(ctx)
	log.Info("DeleteUserURLs was called")
	return nil
}

func TestShortenURL(t *testing.T) {
	testCfg := &config.Config{
		BaseURL: "http://localhost:8081",
	}
	tests := []struct {
		name                 string
		requestBody          string
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			name:                 "Test ShortenURL success",
			requestBody:          "https://practicum.yandex.ru/",
			expectedStatus:       http.StatusCreated,
			expectedResponseBody: testCfg.BaseURL + "/" + utils.GenerateShortURL("https://practicum.yandex.ru/"),
		},
		{
			name:                 "Test ShortenURL without body",
			requestBody:          "",
			expectedStatus:       http.StatusBadRequest,
			expectedResponseBody: "",
		},
	}

	testStorage1 := newTestStorage()
	testHandler1 := NewHandlerURL(testStorage1, testCfg.BaseURL)

	testlog1 := logger.NewLogger()
	ctxLogger := logger.ContextWithLogger(context.Background(), &testlog1)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest("POST", "/shorten", body)
			request = request.WithContext(context.WithValue(ctxLogger, auth.UserIDKey, "test_user_id"))

			w := httptest.NewRecorder()
			testHandler1.ShortenURL(w, request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			resultBody := w.Body.String()
			assert.Equal(t, tt.expectedResponseBody, resultBody)
		})
	}
}

func TestRedirectURL(t *testing.T) {
	testCfg := &config.Config{
		BaseURL: "http://localhost:8081",
	}

	testStorage2 := newTestStorage()
	testHandler2 := NewHandlerURL(testStorage2, testCfg.BaseURL)

	testlog2 := logger.NewLogger()
	ctxLogger := logger.ContextWithLogger(context.Background(), &testlog2)

	r := chi.NewRouter()
	r.HandleFunc("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		req := r.WithContext(context.WithValue(ctxLogger, auth.UserIDKey, "test_user_id"))
		testHandler2.RedirectURL(rw, req)
	})

	testURL := "https://practicum.yandex.ru/"
	shortURL := utils.GenerateShortURL(testURL)
	userID := "test_user_id"
	testStorage2.ShortenURL(context.TODO(), userID, testURL)

	tests := []struct {
		name             string
		URL              string
		router           http.Handler
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "ShortURL is in UrlMap",
			router:           r,
			URL:              shortURL,
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: testURL,
		},
		{
			name:             "ShortURL not in UrlMap",
			router:           r,
			URL:              "invalidURL",
			expectedStatus:   http.StatusGone,
			expectedLocation: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+tt.URL, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.Equal(t, tt.expectedLocation, rr.Header().Get("Location"))
		})
	}
}

func TestJSONShortenURL(t *testing.T) {
	testCfg := &config.Config{
		BaseURL: "http://localhost:8081",
	}
	tests := []struct {
		name                 string
		requestBody          string
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			name:                 "Test JSONShortenURL success.",
			requestBody:          `{"url": "https://practicum.yandex.ru/"}`,
			expectedStatus:       http.StatusCreated,
			expectedResponseBody: `{"result":"` + testCfg.BaseURL + "/" + utils.GenerateShortURL("https://practicum.yandex.ru/") + `"}`,
		},
		{
			name:                 "Test JSONShortenURL invalid JSON",
			requestBody:          `invalid JSON`,
			expectedStatus:       http.StatusBadRequest,
			expectedResponseBody: "",
		},
	}
	testStorage3 := newTestStorage()
	testHandler3 := NewHandlerURL(testStorage3, testCfg.BaseURL)

	testlog3 := logger.NewLogger()
	ctxLogger := logger.ContextWithLogger(context.Background(), &testlog3)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest("POST", "/api/shorten", body)
			request = request.WithContext(context.WithValue(ctxLogger, auth.UserIDKey, "test_user_id"))

			w := httptest.NewRecorder()
			testHandler3.JSONShortenURL(w, request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			resultBody := strings.TrimSpace(w.Body.String())
			assert.Equal(t, tt.expectedResponseBody, resultBody)
		})
	}
}
