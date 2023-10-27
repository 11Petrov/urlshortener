package migrations

import (
	"context"
	"database/sql"
	"time"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upShortener, downShortener)
}

func upShortener(ctx context.Context, tx *sql.Tx) error {
	ctrl, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	query := `
	CREATE TABLE IF NOT EXISTS shortener (
		id SERIAL PRIMARY KEY,
		short_url TEXT NOT NULL,
		original_url TEXT NOT NULL
	);
	
	ALTER TABLE shortener ADD COLUMN IF NOT EXISTS user_id VARCHAR; 
	
    CREATE UNIQUE INDEX IF NOT EXISTS original_url_unique ON shortener(original_url);
	`
	_, err := tx.ExecContext(ctrl, query)
	if err != nil {
		return err
	}
	return nil
}

func downShortener(ctx context.Context, tx *sql.Tx) error {
	ctrl, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	query := `DROP TABLE shortener;`
	_, err := tx.ExecContext(ctrl, query)
	if err != nil {
		return err
	}
	return nil
}
