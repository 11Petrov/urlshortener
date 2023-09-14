package storage

import (
	"errors"

	"github.com/11Petrov/urlshortener/internal/utils"
)

type URLStorage interface {
	SetURL(originalURL string) string
	GetURL(shortURL string) (string, bool)
}

type URLStorageMap struct {
	URLMap map[string]string
}

func NewStorageURLMap() *URLStorageMap {
	return &URLStorageMap{
		URLMap: make(map[string]string),
	}
}

func (s *URLStorageMap) SetURL(originalURL string) string {
	shortURL := utils.GenerateShortURL(originalURL)
	s.URLMap[shortURL] = originalURL
	return shortURL
}

func (s *URLStorageMap) GetURL(shortURL string) (string, error) {
	url, ok := s.URLMap[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}
