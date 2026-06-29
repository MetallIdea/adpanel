CREATE TABLE web_services (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL REFERENCES web_sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    port INT NOT NULL
);