package webservice

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

func (r *Repository) GetAll(ctx context.Context) ([]WebService, error) {
	rows, err := r.db.Query(
		ctx,
		"SELECT id, site_id, name, port FROM web_services",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []WebService

	for rows.Next() {
		var service WebService

		err := rows.Scan(
			&service.ID,
			&service.SiteID,
			&service.Name,
			&service.Port,
		)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}

func (r *Repository) Create(ctx context.Context, service *WebService) error {
	return r.db.QueryRow(
		ctx,
		"INSERT INTO web_services(site_id, name, port) VALUES($1, $2, $3) RETURNING id",
		service.SiteID,
		service.Name,
		service.Port,
	).Scan(&service.ID)
}

func (r *Repository) GetBySiteID(ctx context.Context, siteID int64) ([]WebService, error) {
	rows, err := r.db.Query(
		ctx,
		"SELECT id, site_id, name, port FROM web_services WHERE site_id=$1",
		siteID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []WebService

	for rows.Next() {
		var service WebService

		err := rows.Scan(
			&service.ID,
			&service.SiteID,
			&service.Name,
			&service.Port,
		)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}
