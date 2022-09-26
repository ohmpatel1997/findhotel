package model

import (
	"context"
	"testing"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/ohmpatel1997/findhotel/lib/db/mocks"
	"github.com/ohmpatel1997/findhotel/lib/router"
	"github.com/stretchr/testify/assert"
)

func TestFindDataByIP(t *testing.T) {
	cases := []struct {
		Name          string
		Ip            string
		Resp          *Geolocation
		PopulateDb    func(db *pg.DB) error
		ExpectedError error
	}{
		{
			Name: "success",
			Ip:   "123",
			PopulateDb: func(db *pg.DB) error {
				geo := &Geolocation{
					IP:           "123",
					Country:      "India",
					CountryCode:  "IN",
					City:         "Mumbai",
					Latitude:     "15.323",
					Longitude:    "145.244",
					MysteryValue: "Mumbai",
					CreatedAt:    time.Now(),
					ModifiedAt:   time.Now(),
				}
				_, err := db.Model(geo).Insert()
				return err
			},
			Resp: &Geolocation{
				IP:           "123",
				Country:      "India",
				CountryCode:  "IN",
				City:         "Mumbai",
				Latitude:     "15.323",
				Longitude:    "145.244",
				MysteryValue: "Mumbai",
			},
			ExpectedError: nil,
		},
		{
			Name: "not found",
			Ip:   "1235",
			PopulateDb: func(db *pg.DB) error {
				return nil
			},
			Resp:          nil,
			ExpectedError: router.NewHttpError("data not found with given ip", 404),
		},
	}

	pool, resource := mocks.NewPGContainer(t)
	defer mocks.CloseContainer(t, pool, resource)
	db := mocks.NewDB(t, pool, resource)
	defer db.Close()
	var geo Geolocation

	err := db.Model(&geo).CreateTable(&orm.CreateTableOptions{FKConstraints: true})
	if err != nil {
		t.Fatalf("Error creating schema %v", err)
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)
			err := tt.PopulateDb(db)
			assert.Equal(nil, err)

			modelManager := NewGeoLocationManager(db)
			resp, err := modelManager.FindDataByIP(context.TODO(), tt.Ip)
			if tt.ExpectedError != nil {
				assert.Equal(tt.ExpectedError, err)
				return
			}
			assert.Equal(tt.Resp.IP, resp.IP)
			assert.Equal(tt.Resp.City, resp.City)
			assert.Equal(tt.Resp.CountryCode, resp.CountryCode)
			assert.Equal(tt.Resp.Country, resp.Country)
			assert.Equal(tt.Resp.Longitude, resp.Longitude)
			assert.Equal(tt.Resp.Latitude, resp.Latitude)
			assert.Equal(tt.Resp.MysteryValue, resp.MysteryValue)
		})
	}
}
