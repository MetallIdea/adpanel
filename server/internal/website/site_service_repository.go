package website

import (
	"context"
	"database/sql"
)

type SiteService struct {
	ID       int64  `json:"id"`
	SiteID   int64  `json:"site_id"`
	ServiceID int64  `json:"service_id"`
	Path     string `json:"path"`
}

type SiteServiceRepository struct {
	db *sql.DB
}

func NewSiteServiceRepository(db *sql.DB) *SiteServiceRepository {
	return &SiteServiceRepository{db: db}
}

func (r *SiteServiceRepository) GetBySiteID(ctx context.Context, siteID int64) ([]SiteService, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, site_id, service_id, path FROM site_services WHERE site_id=?",
		siteID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var siteServices []SiteService

	for rows.Next() {
		var ss SiteService
		err := rows.Scan(&ss.ID, &ss.SiteID, &ss.ServiceID, &ss.Path)
		if err != nil {
			return nil, err
		}
		siteServices = append(siteServices, ss)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return siteServices, nil
}

func (r *SiteServiceRepository) Create(ctx context.Context, ss *SiteService) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO site_services(site_id, service_id, path) VALUES(?, ?, ?)",
		ss.SiteID,
		ss.ServiceID,
		ss.Path,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	ss.ID = id
	return nil
}

func (r *SiteServiceRepository) Update(ctx context.Context, ss *SiteService) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE site_services SET site_id=?, service_id=?, path=? WHERE id=?",
		ss.SiteID,
		ss.ServiceID,
		ss.Path,
		ss.ID,
	)
	return err
}

func (r *SiteServiceRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM site_services WHERE id=?",
		id,
	)
	return err
}

func (r *SiteServiceRepository) DeleteBySiteID(ctx context.Context, siteID int64) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM site_services WHERE site_id=?",
		siteID,
	)
	return err
}
