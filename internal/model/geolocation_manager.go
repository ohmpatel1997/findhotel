package model

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/ohmpatel1997/findhotel/lib/router"
)

//go:generate mockery --name GeoLocationManager --output=mocks
type GeoLocationManager interface {
	FindDataByIP(ctx context.Context, ip string) (*Geolocation, error)
	BulkInsert(ctx context.Context, geolocation []*Geolocation) error
}

type manager struct {
	db *pg.DB
}

func NewGeoLocationManager(db *pg.DB) GeoLocationManager {
	return &manager{
		db: db,
	}
}

func (m *manager) FindDataByIP(ctx context.Context, ip string) (*Geolocation, error) {
	var resp Geolocation

	if len(ip) == 0 {
		return nil, router.NewHttpError("invalid ip", 400)
	}

	err := m.db.ModelContext(ctx, &resp).Where("ip = ?", ip).Select()
	switch {
	case errors.Is(err, pg.ErrNoRows):
		return nil, router.NewHttpError("data not found with given ip", 404)
	case err != nil:
		return nil, router.NewHttpError(err.Error(), 500)
	}
	return &resp, nil
}

func (m *manager) BulkInsert(ctx context.Context, geolocation []*Geolocation) error {
	return m.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		_, err := tx.Model(&geolocation).Insert()
		return err
	})
}
