package storage

import (
	"errors"

	"github.com/11Petrov/urlshortener/internal/utils"
)

// URLStore определяет интерфейс для хранилища URL
type URLStore interface {
	ShortenURL(originalURL string) string
	RedirectURL(shortURL string) (string, error)
}

// RepoURL - структура, реализующая интерфейс URLStore
type repoURL struct {
	URLMap map[string]string
}

// NewRepoURL создает новый экземпляр RepoURL
func NewRepoURL() URLStore {
	return &repoURL{
		URLMap: make(map[string]string),
	}
}

// ShortenURL сокращает оригинальный URL и сохраняет его в хранилище, возвращая сокращенный URL
func (s *repoURL) ShortenURL(originalURL string) string {
	shortURL := utils.GenerateShortURL(originalURL)
	s.URLMap[shortURL] = originalURL
	return shortURL
}

// RedirectURL возвращает оригинальный URL
func (s *repoURL) RedirectURL(shortURL string) (string, error) {
	url, ok := s.URLMap[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}
