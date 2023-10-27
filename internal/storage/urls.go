package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/models"
	"github.com/11Petrov/urlshortener/internal/utils"
)

// URLStore определяет интерфейс для хранилища URL
type URLStore interface {
	ShortenURL(ctx context.Context, userID, originalURL string) (string, error)
	RedirectURL(ctx context.Context, userID, shortURL string) (string, error)
	Ping(ctx context.Context) error
	BatchShortenURL(ctx context.Context, userID, originalURL string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]models.Event, error)
}

// RepoURL - структура, реализующая интерфейс URLStore
type repoURL struct {
	URLMap  map[string]map[string]string
	file    *os.File
	encoder *json.Encoder
}

func NewRepo(cfg *config.Config, ctx context.Context) URLStore {
	log := logger.LoggerFromContext(ctx)
	if cfg.DatabaseAddress != "" {
		store, err := NewDBStore(cfg.DatabaseAddress, ctx)
		if err != nil {
			log.Fatal(err)
		}
		return store
	} else {
		store, err := NewRepoURL(cfg.FilePath, ctx)
		if err != nil {
			log.Fatal(err)
		}
		return store
	}
}

// NewRepoURL создает новый экземпляр RepoURL
func NewRepoURL(filename string, ctx context.Context) (URLStore, error) {
	log := logger.LoggerFromContext(ctx)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("error OpneFile %s", err)
		return nil, err
	}

	decoder := json.NewDecoder(file)
	URLMap := make(map[string]map[string]string)
	for {
		var event models.Event
		if err := decoder.Decode(&event); err != nil {
			log.Errorf("error Decode to event %s", err)
			break
		}
		if _, ok := URLMap[event.UserID]; !ok {
			URLMap[event.UserID] = make(map[string]string)
		}
		URLMap[event.UserID][event.ShortURL] = event.OriginalURL
	}

	return &repoURL{
		URLMap:  URLMap,
		file:    file,
		encoder: json.NewEncoder(file),
	}, err
}

// ShortenURL сокращает оригинальный URL и сохраняет его в хранилище, возвращая сокращенный URL
func (r *repoURL) ShortenURL(ctx context.Context, userID, originalURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	shortURL := utils.GenerateShortURL(originalURL)
	if _, ok := r.URLMap[userID]; !ok {
		r.URLMap[userID] = make(map[string]string)
	}
	r.URLMap[userID][shortURL] = originalURL

	event := models.Event{
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	data, err := json.Marshal(&event)
	if err != nil {
		log.Errorf("error json.Marshal(&event) %s", err)
		return "", err
	}

	_, err = r.file.Write(append(data, '\n'))
	if err != nil {
		log.Errorf("error Write %s", err)
		return "", err
	}
	r.file.Sync()
	return shortURL, nil
}

// RedirectURL возвращает оригинальный URL
func (r *repoURL) RedirectURL(ctx context.Context, userID, shortURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	url, ok := r.URLMap[userID][shortURL]
	if !ok {
		log.Error("error URLMap[shortURL]")
		return "", errors.New("url not found")
	}
	return url, nil
}

func (r *repoURL) BatchShortenURL(ctx context.Context, userID, originalURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	log.Info("BatchShortenURL function was called")
	return "", nil
}

func (r *repoURL) Ping(ctx context.Context) error {
	log := logger.LoggerFromContext(ctx)
	log.Info("Ping function was called")
	return nil
}

func (r *repoURL) GetUserURLs(ctx context.Context, userID string) ([]models.Event, error) {
	log := logger.LoggerFromContext(ctx)

	urls, ok := r.URLMap[userID]
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
