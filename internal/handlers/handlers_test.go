package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/11Petrov/urlshortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURL(t *testing.T) {

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}

	tests := []struct {
		name         string
		requestBody  string
		expectedCode int
		args         args
	}{
		{
			name:         "Test ShortenURL success",
			requestBody:  "https://practicum.yandex.ru/",
			expectedCode: http.StatusCreated,
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/", strings.NewReader("https://practicum.yandex.ru/")),
			},
		},
		{
			name:         "Test ShortenURL without body",
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ShortenURL(tt.args.w, tt.args.r)

			result := tt.args.w.Result()
			assert.Equal(t, result.StatusCode, tt.expectedCode)

			defer result.Body.Close()

			if len(tt.requestBody) > 0 {
				resultBody, err := io.ReadAll(result.Body)
				require.NoError(t, err)

				shortURL := storage.HostURL + GenerateShortURL(tt.requestBody)
				assert.Equal(t, string(resultBody), shortURL)
			}
		})
	}
}

func TestRedirectURL(t *testing.T) {
	storage.URLMap = make(map[string]string)

	r := chi.NewRouter()
	r.HandleFunc("/{id}", RedirectURL)

	testURL := "https://practicum.yandex.ru/"
	shortURL := GenerateShortURL(testURL)
	storage.URLMap[shortURL] = testURL

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
