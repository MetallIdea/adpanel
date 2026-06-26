package website

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]Website, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, url FROM web_sites")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []Website

	for rows.Next() {
		var site Website

		err := rows.Scan(&site.ID, &site.Name, &site.URL)
		if err != nil {
			return nil, err
		}

		sites = append(sites, site)
	}

	return sites, nil
}
