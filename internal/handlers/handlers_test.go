package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/storage"
	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStorage struct {
	URLMap map[string]string
}

func newTestStorage() storage.DataStore {
	return &testStorage{
		URLMap: make(map[string]string),
	}
}

func (t *testStorage) ShortenURL(originalURL string) string {
	shortURL := utils.GenerateShortURL(originalURL)
	t.URLMap[shortURL] = originalURL
	return shortURL
}

func (t *testStorage) RedirectURL(shortURL string) (string, error) {
	url, ok := t.URLMap[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}

func TestShortenURL(t *testing.T) {
	testCfg := &config.Config{
		ServerAddress: "localhost:8081",
		BaseURL:       "http://localhost:8081",
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
			expectedResponseBody: "Request body is missing\n",
		},
	}

	testStorage1 := newTestStorage()
	testHandler := NewURLHandler(testStorage1, testCfg.BaseURL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest("POST", "/shorten", body)
			w := httptest.NewRecorder()
			testHandler.ShortenURL(w, request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			resultBody := w.Body.String()
			assert.Equal(t, tt.expectedResponseBody, resultBody)
		})
	}
}

func TestRedirectURL(t *testing.T) {
	testCfg := &config.Config{
		ServerAddress: "localhost:8081",
		BaseURL:       "http://localhost:8081",
	}

	testStorage2 := storage.NewStorageURL()
	testHandler2 := NewURLHandler(testStorage2, testCfg.BaseURL)

	r := chi.NewRouter()
	r.HandleFunc("/{id}", testHandler2.RedirectURL)

	testURL := "https://practicum.yandex.ru/"
	shortURL := utils.GenerateShortURL(testURL)
	testStorage2.ShortenURL(testURL)

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
