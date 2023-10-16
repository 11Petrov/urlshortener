package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/11Petrov/urlshortener/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Database struct {
	db *sql.DB
}

func NewDBStore(databaseAddress string) (*Database, error) {
	db, err := sql.Open("pgx", databaseAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect")
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	q := `CREATE TABLE IF NOT EXISTS shortener (
		id SERIAL PRIMARY KEY,
		short_url TEXT NOT NULL,
		original_url TEXT NOT NULL
	);`
	_, err = db.Exec(q)
	if err != nil {
		return nil, err
	}
	uniqueIndex := "CREATE UNIQUE INDEX IF NOT EXISTS original_url_unique ON shortener(original_url)"
	_, err = db.Exec(uniqueIndex)
	if err != nil {
		return nil, err
	}
	d := &Database{
		db: db,
	}
	return d, nil
}

func (s *Database) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	shortURL := utils.GenerateShortURL(originalURL)

	c, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	_, err := s.db.ExecContext(c, `INSERT INTO shortener(short_url, original_url) VALUES($1, $2)`, shortURL, originalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var res string
			s.db.QueryRowContext(c, `SELECT short_url FROM shortener WHERE original_url = $1`, originalURL).Scan(&res)
			return res, err
		}
		return "", err
	}

	return shortURL, err
}

func (s *Database) RedirectURL(ctx context.Context, shortURL string) (string, error) {
	var originalURL string
	c, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	row := s.db.QueryRowContext(c, `SELECT original_url FROM shortener WHERE short_url = $1`, shortURL)
	if err := row.Scan(&originalURL); err != nil {
		return "", nil
	}
	return originalURL, nil
}

func (s *Database) Ping(ctx context.Context) error {
	err := s.db.PingContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Database) BatchShortenURL(ctx context.Context, originalURL string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	shortURL := utils.GenerateShortURL(originalURL)
	_, err = tx.ExecContext(ctx, "INSERT INTO shortener(short_url, original_url) VALUES($1, $2)", shortURL, originalURL)
	if err != nil {
		return "", err
	}
	return shortURL, tx.Commit()
}
