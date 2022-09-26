package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohmpatel1997/findhotel/internal/model"
	modelMocks "github.com/ohmpatel1997/findhotel/internal/model/mocks"
	"github.com/ohmpatel1997/findhotel/lib/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGeolocation(t *testing.T) {
	cases := []struct {
		Name          string
		Req           *GetRequest
		ExpectedResp  *GeoLocationResponse
		ExpectedError error
		MocksInit     func() *modelMocks.GeoLocationManager
	}{
		{
			Name: "Success",
			Req:  &GetRequest{IP: "ip1"},
			ExpectedResp: &GeoLocationResponse{
				IP:           "ip1",
				Country:      "india",
				CountryCode:  "IN",
				City:         "mumbai",
				Latitude:     "12.2344",
				Longitude:    "149.3123123",
				MysteryValue: "MUMbai",
			},
			MocksInit: func() *modelMocks.GeoLocationManager {
				manager := new(modelMocks.GeoLocationManager)
				manager.On("FindDataByIP", mock.Anything, "ip1").Return(&model.Geolocation{
					ID:           uuid.New(),
					IP:           "ip1",
					Country:      "india",
					CountryCode:  "IN",
					City:         "mumbai",
					Latitude:     "12.2344",
					Longitude:    "149.3123123",
					MysteryValue: "MUMbai",
					CreatedAt:    time.Now(),
					ModifiedAt:   time.Now(),
				}, true, nil)
				return manager
			},
			ExpectedError: nil,
		},
		{
			Name:         "400 bad request",
			Req:          &GetRequest{IP: ""},
			ExpectedResp: nil,
			MocksInit: func() *modelMocks.GeoLocationManager {
				return nil
			},
			ExpectedError: router.NewHttpError("invalid ip", 400),
		},
		{
			Name:         "404 not found",
			Req:          &GetRequest{IP: "ip1"},
			ExpectedResp: nil,
			MocksInit: func() *modelMocks.GeoLocationManager {
				manager := new(modelMocks.GeoLocationManager)
				manager.On("FindDataByIP", mock.Anything, "ip1").Return(nil, false, nil)
				return manager
			},
			ExpectedError: router.NewHttpError("not found", 404),
		},
		{
			Name:         "500 internal error",
			Req:          &GetRequest{IP: "ip1"},
			ExpectedResp: nil,
			MocksInit: func() *modelMocks.GeoLocationManager {
				manager := new(modelMocks.GeoLocationManager)
				manager.On("FindDataByIP", mock.Anything, "ip1").Return(nil, false, errors.New("custom error"))
				return manager
			},
			ExpectedError: router.NewHttpError("custom error", 500),
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)
			srv := NewGeolocationService(tt.MocksInit())
			resp, err := srv.GetIPData(context.TODO(), tt.Req)
			assert.Equal(tt.ExpectedResp, resp)
			assert.Equal(tt.ExpectedError, err)
		})
	}
}
