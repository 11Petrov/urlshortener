-- +goose Up
CREATE TABLE IF NOT EXISTS shortener (
    id SERIAL PRIMARY KEY,
    short_url TEXT NOT NULL,
    original_url TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS original_url_unique ON shortener(original_url);

-- +goose Down
DROP TABLE shortener;