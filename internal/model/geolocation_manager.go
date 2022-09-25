package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-pg/pg/v10"
)

type GeoLocationManager interface {
	FindDataByIP(ctx context.Context, ip string) (*Geolocation, bool, error)
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

func (m *manager) FindDataByIP(ctx context.Context, ip string) (*Geolocation, bool, error) {
	var resp Geolocation

	if len(ip) == 0 {
		return nil, false, fmt.Errorf("ip can not be empty")
	}

	err := m.db.ModelContext(ctx, &resp).Where("ip = ?", ip).Select()
	switch {
	case errors.Is(err, pg.ErrNoRows):
		return nil, false, fmt.Errorf("data not found with given ip")
	case err != nil:
		return nil, false, err
	}
	return &resp, true, nil
}

func (m *manager) BulkInsert(ctx context.Context, geolocation []*Geolocation) error {
	return m.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		_, err := tx.Model(geolocation).Insert()
		return err
	})
}
