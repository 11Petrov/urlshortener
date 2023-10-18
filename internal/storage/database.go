package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/11Petrov/urlshortener/internal/logger"
	e "github.com/11Petrov/urlshortener/internal/storage/errors"
	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
)

type Database struct {
	db *sql.DB
}

func NewDBStore(databaseAddress string, ctx context.Context) (*Database, error) {
	log := logger.LoggerFromContext(ctx)
	db, err := sql.Open("pgx", databaseAddress)
	if err != nil {
		log.Errorf("failed to connect %s", err)
		return nil, fmt.Errorf("failed to connect")
	}
	err = db.Ping()
	if err != nil {
		log.Errorf("error Ping() %s", err)
		return nil, err
	}

	log.Info("Start migrating database")
	err = goose.Up(db, ".")
	if err != nil {
		log.Errorf("error goose.Up %s", err)
		return nil, err
	}

	d := &Database{
		db: db,
	}
	return d, nil
}

func (s *Database) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	shortURL := utils.GenerateShortURL(originalURL)

	c, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	_, err := s.db.ExecContext(c, `INSERT INTO shortener(short_url, original_url) VALUES($1, $2)`, shortURL, originalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var res string
			s.db.QueryRowContext(c, `SELECT short_url FROM shortener WHERE original_url = $1`, originalURL).Scan(&res)
			return res, e.ErrUnique
		}
		log.Errorf("error ExecContext %s", err)
		return "", err
	}

	return shortURL, err
}

func (s *Database) RedirectURL(ctx context.Context, shortURL string) (string, error) {
	var originalURL string
	log := logger.LoggerFromContext(ctx)
	c, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	row := s.db.QueryRowContext(c, `SELECT original_url FROM shortener WHERE short_url = $1`, shortURL)
	if err := row.Scan(&originalURL); err != nil {
		log.Errorf("error QueryRowContext %s", err)
		return "", nil
	}
	return originalURL, nil
}

func (s *Database) Ping(ctx context.Context) error {
	log := logger.LoggerFromContext(ctx)
	err := s.db.PingContext(ctx)
	if err != nil {
		log.Errorf("error PingContext %s", err)
		return err
	}
	return nil
}

func (s *Database) BatchShortenURL(ctx context.Context, originalURL string) (string, error) {
	log := logger.LoggerFromContext(ctx)
	tx, err := s.db.Begin()
	if err != nil {
		log.Errorf("error Begin() %s", err)
		return "", err
	}
	defer tx.Rollback()

	shortURL := utils.GenerateShortURL(originalURL)
	_, err = tx.ExecContext(ctx, "INSERT INTO shortener(short_url, original_url) VALUES($1, $2)", shortURL, originalURL)
	if err != nil {
		log.Errorf("error ExecContext %s", err)
		return "", err
	}
	return shortURL, tx.Commit()
}