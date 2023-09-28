package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/google/uuid"
)

// URLStore определяет интерфейс для хранилища URL
type URLStore interface {
	ShortenURL(originalURL string) (string, error)
	RedirectURL(shortURL string) (string, error)
}

// Event представляет информацию о сокращенном URL для сохранения в файле
type Event struct {
	ID          uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

// RepoURL - структура, реализующая интерфейс URLStore
type repoURL struct {
	URLMap  map[string]string
	file    *os.File
	encoder *json.Encoder
}

// NewRepoURL создает новый экземпляр RepoURL
func NewRepoURL(filename string) (URLStore, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	URLMap := make(map[string]string)
	for {
		var event Event
		if err := decoder.Decode(&event); err != nil {
			break
		}
		URLMap[event.ShortURL] = event.OriginalURL
	}

	return &repoURL{
		URLMap:  URLMap,
		file:    file,
		encoder: json.NewEncoder(file),
	}, err
}

// ShortenURL сокращает оригинальный URL и сохраняет его в хранилище, возвращая сокращенный URL
func (r *repoURL) ShortenURL(originalURL string) (string, error) {
	shortURL := utils.GenerateShortURL(originalURL)
	r.URLMap[shortURL] = originalURL

	event := Event{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	data, err := json.Marshal(&event)
	if err != nil {
		return "", err
	}

	_, err = r.file.Write(append(data, '\n'))
	if err != nil {
		return "", err
	}
	r.file.Sync()
	return shortURL, nil
}

// RedirectURL возвращает оригинальный URL
func (r *repoURL) RedirectURL(shortURL string) (string, error) {
	url, ok := r.URLMap[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}
