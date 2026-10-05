package website

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]Website, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, url FROM web_sites")
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

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, name, url FROM web_sites WHERE id=?",
		id,
	).Scan(&site.ID, &site.Name, &site.URL)

	if err != nil {
		return nil, err
	}

	return &site, nil
}

func (r *Repository) Create(ctx context.Context, site *Website) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO web_sites(name, url) VALUES(?, ?)",
		site.Name,
		site.URL,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	site.ID = id
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM web_sites WHERE id=?",
		id,
	)

	return err
}

func (r *Repository) Update(ctx context.Context, site *Website) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE web_sites SET name=?, url=? WHERE id=?",
		site.Name,
		site.URL,
		site.ID,
	)

	return err
}
