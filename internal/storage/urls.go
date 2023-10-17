package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/11Petrov/urlshortener/cmd/config"
	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// URLStore определяет интерфейс для хранилища URL
type URLStore interface {
	ShortenURL(ctx context.Context, originalURL string) (string, error)
	RedirectURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	BatchShortenURL(ctx context.Context, originalURL string) (string, error)
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
	log     zap.SugaredLogger
}

func NewRepo(cfg *config.Config, log zap.SugaredLogger) URLStore {
	if cfg.DatabaseAddress != "" {
		store, err := NewDBStore(cfg.DatabaseAddress, log)
		if err != nil {
			log.Fatal(err)
		}
		return store
	} else {
		store, err := NewRepoURL(cfg.FilePath, log)
		if err != nil {
			log.Fatal(err)
		}
		return store
	}
}

// NewRepoURL создает новый экземпляр RepoURL
func NewRepoURL(filename string, log zap.SugaredLogger) (URLStore, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("error OpneFile %s", err)
		return nil, err
	}

	decoder := json.NewDecoder(file)
	URLMap := make(map[string]string)
	for {
		var event Event
		if err := decoder.Decode(&event); err != nil {
			log.Errorf("error Decode to event %s", err)
			break
		}
		URLMap[event.ShortURL] = event.OriginalURL
	}

	return &repoURL{
		URLMap:  URLMap,
		file:    file,
		encoder: json.NewEncoder(file),
		log:     log,
	}, err
}

// ShortenURL сокращает оригинальный URL и сохраняет его в хранилище, возвращая сокращенный URL
func (r *repoURL) ShortenURL(_ context.Context, originalURL string) (string, error) {
	shortURL := utils.GenerateShortURL(originalURL)
	r.URLMap[shortURL] = originalURL

	event := Event{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	data, err := json.Marshal(&event)
	if err != nil {
		r.log.Errorf("error json.Marshal(&event) %s", err)
		return "", err
	}

	_, err = r.file.Write(append(data, '\n'))
	if err != nil {
		r.log.Errorf("error Write %s", err)
		return "", err
	}
	r.file.Sync()
	return shortURL, nil
}

// RedirectURL возвращает оригинальный URL
func (r *repoURL) RedirectURL(_ context.Context, shortURL string) (string, error) {
	url, ok := r.URLMap[shortURL]
	if !ok {
		r.log.Error("error URLMap[shortURL]")
		return "", errors.New("url not found")
	}
	return url, nil
}

func (r *repoURL) Ping(_ context.Context) error {
	return nil
}

func (r *repoURL) BatchShortenURL(_ context.Context, originalURL string) (string, error) {
	return "", nil
}
