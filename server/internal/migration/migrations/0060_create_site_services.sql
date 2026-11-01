-- +goose Up
CREATE TABLE IF NOT EXISTS site_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id INTEGER NOT NULL REFERENCES web_sites(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES web_services(id) ON DELETE CASCADE,
    path TEXT NOT NULL DEFAULT '/',
    UNIQUE(site_id, service_id)
);

-- +goose Down
DROP TABLE IF EXISTS site_services;
