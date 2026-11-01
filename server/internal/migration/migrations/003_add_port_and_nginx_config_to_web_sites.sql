-- +goose Up
ALTER TABLE web_sites ADD COLUMN port INTEGER NOT NULL DEFAULT 8080;

-- +goose Down
ALTER TABLE web_sites DROP COLUMN port;
