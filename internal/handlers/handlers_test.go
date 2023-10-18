package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestURLStore interface {
	ShortenURL(ctx context.Context, originalURL string) (string, error)
	RedirectURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	BatchShortenURL(ctx context.Context, originalURL string) (string, error)
}

type testStorage struct {
	URLMap map[string]string
}

func newTestStorage() TestURLStore {
	return &testStorage{
		URLMap: make(map[string]string),
	}
}

func (t *testStorage) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	shortURL := utils.GenerateShortURL(originalURL)
	t.URLMap[shortURL] = originalURL
	return shortURL, nil
}

func (t *testStorage) RedirectURL(ctx context.Context, shortURL string) (string, error) {
	url, ok := t.URLMap[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}

// BatchShortenURL implements URLStore.
func (t *testStorage) BatchShortenURL(ctx context.Context, originalURL string) (string, error) {
	panic("unimplemented")
}

// Ping implements URLStore.
func (t *testStorage) Ping(ctx context.Context) error {
	panic("unimplemented")
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

	test_log1 := logger.NewLogger()
	ctxLogger := logger.ContextWithLogger(context.Background(), &test_log1)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest("POST", "/shorten", body)
			request = request.WithContext(ctxLogger)

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

	test_log2 := logger.NewLogger()
	ctxLogger := logger.ContextWithLogger(context.Background(), &test_log2)

	r := chi.NewRouter()
	r.HandleFunc("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		req := r.WithContext(ctxLogger)
		testHandler2.RedirectURL(rw, req)
	})

	testURL := "https://practicum.yandex.ru/"
	shortURL := utils.GenerateShortURL(testURL)
	testStorage2.ShortenURL(context.TODO(), testURL)

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
			expectedStatus:   http.StatusBadRequest,
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

	test_log2 := logger.NewLogger()
	ctxLogger := logger.ContextWithLogger(context.Background(), &test_log2)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			fmt.Println(body)
			request := httptest.NewRequest("POST", "/api/shorten", body)
			request = request.WithContext(ctxLogger)

			w := httptest.NewRecorder()
			testHandler3.JSONShortenURL(w, request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			resultBody := strings.TrimSpace(w.Body.String())
			assert.Equal(t, tt.expectedResponseBody, resultBody)
		})
	}
}
