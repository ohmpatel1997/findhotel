package controller

import (
	"net/http"

	"github.com/ohmpatel1997/findhotel/internal/service"
)

const (
	clientApiVersion = "v1"
)

type ClientController interface {

	//metadata
	GetAPIVersion() string
	GetAPIVersionPath(string) string

	GetGeolocationData(http.ResponseWriter, *http.Request)
}

type clientController struct {
	geolocationSrv service.GeoLocationService
}

func NewController(geolocation service.GeoLocationService) ClientController {
	return &clientController{
		geolocationSrv: geolocation,
	}
}

func (c *clientController) GetAPIVersion() string {
	return clientApiVersion
}

func (c *clientController) GetAPIVersionPath(p string) string {
	return "/" + clientApiVersion + p
}
