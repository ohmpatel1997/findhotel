// Package pgsql holds the init related functionality
package pgsql

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/ohmpatel1997/findhotel/lib/config"
)

// New database connection to a init database
func New(cfg *config.Database, psn string) (*pg.DB, error) {
	timeout := cfg.Timeout

	if !cfg.SSLMode {
		psn += "?sslmode=disable"
	}

	u, err := pg.ParseURL(psn)
	if err != nil {
		return nil, err
	}
	if timeout > 0 {
		u.ReadTimeout = time.Second * time.Duration(timeout)
		u.WriteTimeout = time.Second * time.Duration(timeout)
	}

	db := pg.Connect(u)

	_, err = db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}

	// extension for auto generate uuid for user table
	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"")
	if err != nil {
		return nil, err
	}

	if timeout > 0 {
		db.WithTimeout(time.Second * time.Duration(timeout))
	}

	return db, nil
}
