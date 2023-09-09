package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/11Petrov/urlshortener/internal/storage"
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

	storage.UrlMap = make(map[string]string)
	storage.UrlMap["EwHXdJfB"] = "https://practicum.yandex.ru/"

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "ShortURL is in UrlMap",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/EwHXdJfB", nil),
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "https://practicum.yandex.ru/",
		},
		{
			name: "ShortURL is not in UrlMap",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/TdJXdJmF", nil),
			},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RedirectURL(tt.args.w, tt.args.r)

			result := tt.args.w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			if tt.expectedHeader != "" {
				header := result.Header.Get("Location")
				assert.Equal(t, tt.expectedHeader, header)
			}

		})
	}
}
