package model

import (
	"time"

	"github.com/google/uuid"
)

type Geolocation struct {
	ID           uuid.UUID `pg:"id, type:uuid, default:gen_random_uuid(), unique"`
	IP           string    `pg:"ip"`
	Country      string    `pg:"country"`
	CountryCode  string    `pg:"country_code"`
	City         string    `pg:"city"`
	Latitude     string    `pg:"latitude"`
	Longitude    string    `pg:"longitude"`
	MysteryValue string    `pg:"mystery_value"`
	CreatedAt    time.Time `sql:"DEFAULT:current_timestamp"`
	ModifiedAt   time.Time `sql:"DEFAULT:current_timestamp"`
}
