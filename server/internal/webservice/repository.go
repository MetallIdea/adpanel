package webservice

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

func (r *Repository) GetAll(ctx context.Context) ([]WebService, error) {
	rows, err := r.db.QueryContext(
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *Repository) Create(ctx context.Context, service *WebService) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO web_services(site_id, name, port) VALUES(?, ?, ?)",
		service.SiteID,
		service.Name,
		service.Port,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	service.ID = id
	return nil
}

func (r *Repository) GetBySiteID(ctx context.Context, siteID int64) ([]WebService, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, site_id, name, port FROM web_services WHERE site_id=?",
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *Repository) GetServiceByID(ctx context.Context, id int64) (*WebService, error) {
	var service WebService

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, site_id, name, port FROM web_services WHERE id=?",
		id,
	).Scan(
		&service.ID,
		&service.SiteID,
		&service.Name,
		&service.Port,
	)

	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *Repository) Update(ctx context.Context, service *WebService) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE web_services SET site_id=?, name=?, port=? WHERE id=?",
		service.SiteID,
		service.Name,
		service.Port,
		service.ID,
	)

	return err
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM web_services WHERE id=?",
		id,
	)

	return err
}
