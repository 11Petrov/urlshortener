package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/11Petrov/urlshortener/internal/logger"
	"github.com/11Petrov/urlshortener/internal/models"
	storageErrors "github.com/11Petrov/urlshortener/internal/storage/errors"
	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

type Database struct {
	db *pgxpool.Pool
}

func NewDBStore(databaseAddress string, ctx context.Context) (*Database, error) {
	log := logger.LoggerFromContext(ctx)

	// Открываем соединение для миграции
	migrationDB, err := sql.Open("pgx", databaseAddress)
	if err != nil {
		log.Errorf("failed to connect for migration: %s", err)
		return nil, err
	}
	defer migrationDB.Close()

	// Проводим миграцию
	log.Info("Start migrating database")
	err = goose.Up(migrationDB, ".")
	if err != nil {
		log.Errorf("error goose.Up: %s", err)
		return nil, err
	}

	// Открываем пул соединений для реальных операций
	db, err := pgxpool.New(context.Background(), databaseAddress)
	if err != nil {
		log.Errorf("failed to connect: %s", err)
		return nil, err
	}

	d := &Database{
		db: db,
	}
	return d, nil
}

func (s *Database) ShortenURL(ctx context.Context, userID, originalURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	shortURL := utils.GenerateShortURL(originalURL)

	_, err := s.db.Exec(ctx, `INSERT INTO shortener(short_url, original_url, user_id) VALUES($1, $2, $3)`, shortURL, originalURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var res string
			s.db.QueryRow(ctx, `SELECT short_url FROM shortener WHERE original_url = $1`, originalURL).Scan(&res)
			return res, storageErrors.ErrUnique
		}
		log.Errorf("error ExecContext %s", err)
		return "", err
	}

	return shortURL, err
}

func (s *Database) RedirectURL(ctx context.Context, userID, shortURL string) (string, error) {
	var originalURL string
	log := logger.LoggerFromContext(ctx)

	row := s.db.QueryRow(ctx, `SELECT original_url FROM shortener WHERE short_url = $1 AND is_deleted = false`, shortURL)
	if err := row.Scan(&originalURL); err != nil {
		log.Errorf("row.Scan error", err)
		return "", err
	}
	return originalURL, nil
}

func (s *Database) Ping(ctx context.Context) error {
	log := logger.LoggerFromContext(ctx)
	err := s.db.Ping(ctx)
	if err != nil {
		log.Errorf("error PingContext %s", err)
		return err
	}
	return nil
}

func (s *Database) BatchShortenURL(ctx context.Context, userID, originalURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Errorf("error Begin() %s", err)
		return "", err
	}
	defer tx.Rollback(ctx)

	shortURL := utils.GenerateShortURL(originalURL)
	_, err = tx.Exec(ctx, "INSERT INTO shortener(short_url, original_url, user_id) VALUES($1, $2, $3)", shortURL, originalURL, userID)
	if err != nil {
		log.Errorf("error ExecContext %s", err)
		return "", err
	}
	return shortURL, tx.Commit(ctx)
}

func (s *Database) GetUserURLs(ctx context.Context, userID string, baseURL string) ([]models.Event, error) {
	log := logger.LoggerFromContext(ctx)
	var events []models.Event

	rows, err := s.db.Query(ctx, `SELECT short_url, original_url FROM shortener WHERE user_id = $1`, userID)
	if err != nil {
		log.Errorf("QueryContext error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ShortURL, &e.OriginalURL); err != nil {
			log.Errorf("Scan error", err)
			return nil, err
		}
		e.ShortURL = baseURL + "/" + e.ShortURL
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("rows.Err()", err)
		return nil, err
	}
	return events, nil
}

func (s *Database) DeleteUserURLs(ctx context.Context, userID string, urls []string) error {
	log := logger.LoggerFromContext(ctx)
	log.Info("Start DeleteUserURLs in database.go")

	batch := pgx.Batch{}
	for _, url := range urls {
		batch.Queue("UPDATE shortener SET is_deleted = true WHERE short_url = $1 AND user_id = $2;", url, userID)
	}

	br := s.db.SendBatch(context.Background(), &batch)
	err := br.Close()
	if err != nil {
		log.Errorf("SendBatch error", err)
		return err
	}

	log.Info("End DeleteUserURLs in database.go")
	return err
}
