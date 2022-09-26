package service

import (
	"context"

	"github.com/ohmpatel1997/findhotel/internal/model"
	"github.com/ohmpatel1997/findhotel/lib/router"
)

//go:generate mockery --name GeoLocationService --output=mocks
type GeoLocationService interface {
	GetIPData(context.Context, *GetRequest) (*GeoLocationResponse, error)
}

type geolocation struct {
	manager model.GeoLocationManager
}

func NewGeolocationService(mn model.GeoLocationManager) GeoLocationService {
	return &geolocation{
		mn,
	}
}

func (g *geolocation) GetIPData(ctx context.Context, request *GetRequest) (*GeoLocationResponse, error) {
	if len(request.IP) == 0 {
		return nil, router.NewHttpError("invalid ip", 400)
	}

	data, found, err := g.manager.FindDataByIP(ctx, request.IP)
	if err != nil {
		return nil, router.NewHttpError(err.Error(), 500)
	}

	if !found {
		return nil, router.NewHttpError("not found", 404)
	}

	return &GeoLocationResponse{
		IP:           data.IP,
		CountryCode:  data.CountryCode,
		Country:      data.Country,
		City:         data.City,
		Latitude:     data.Latitude,
		Longitude:    data.Longitude,
		MysteryValue: data.MysteryValue,
	}, nil
}
