-- +goose Up
CREATE TABLE IF NOT EXISTS web_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id INTEGER NOT NULL REFERENCES web_sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    port INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS web_services;
