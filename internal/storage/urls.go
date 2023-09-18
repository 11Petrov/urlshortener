package storage

import (
	"errors"

	"github.com/11Petrov/urlshortener/internal/utils"
)

type DataStore interface {
	ShortenURL(originalURL string) string
	RedirectURL(shortURL string) (string, error)
}

type StorageURL struct {
	URLMap map[string]string
}

func NewStorageURL() DataStore {
	return &StorageURL{
		URLMap: make(map[string]string),
	}
}

func (s *StorageURL) ShortenURL(originalURL string) string {
	shortURL := utils.GenerateShortURL(originalURL)
	s.URLMap[shortURL] = originalURL
	return shortURL
}

func (s *StorageURL) RedirectURL(shortURL string) (string, error) {
	url, ok := s.URLMap[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}
