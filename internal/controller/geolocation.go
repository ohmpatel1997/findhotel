package controller

import (
	"net/http"

	"github.com/ohmpatel1997/findhotel/internal/service"
	"github.com/ohmpatel1997/findhotel/lib/router"
)

const (
	ParamIP = "ip"
)

func (c *clientController) GetGeolocationData(w http.ResponseWriter, r *http.Request) {
	req := new(service.GetRequest)
	req.IP = r.URL.Query().Get(ParamIP)
	if len(req.IP) == 0 {
		router.RenderError(w, router.NewHttpError("path param could not be found", 400))
		return
	}

	response, err := c.geolocationSrv.GetIPData(r.Context(), req)
	if err != nil {
		router.RenderError(w, err)
		return
	}

	router.RenderJSON(router.Response{
		Writer: w,
		Data:   response,
		Status: 200,
	})
}
