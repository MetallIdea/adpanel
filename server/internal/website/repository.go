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

func (r *Repository) GetByID(ctx context.Context, id int64) (*Website, error) {
	var site Website

	err := r.db.QueryRow(
		ctx,
		"SELECT id, name, url FROM web_sites WHERE id=$1",
		id,
	).Scan(&site.ID, &site.Name, &site.URL)

	if err != nil {
		return nil, err
	}

	return &site, nil
}

func (r *Repository) Create(ctx context.Context, site *Website) error {
	return r.db.QueryRow(
		ctx,
		"INSERT INTO web_sites(name, url) VALUES($1, $2) RETURNING id",
		site.Name,
		site.URL,
	).Scan(&site.ID)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(
		ctx,
		"DELETE FROM web_sites WHERE id=$1",
		id,
	)

	return err
}

func (r *Repository) Update(ctx context.Context, site *Website) error {
	_, err := r.db.Exec(
		ctx,
		"UPDATE web_sites SET name=$1, url=$2 WHERE id=$3",
		site.Name,
		site.URL,
		site.ID,
	)

	return err
}
