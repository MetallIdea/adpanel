-- +goose Up
ALTER TABLE web_sites
ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

-- +goose Down
ALTER TABLE web_sites
DROP COLUMN status;
