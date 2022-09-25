package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ohmpatel1997/findhotel/internal/controller"
	"github.com/ohmpatel1997/findhotel/internal/model"
	"github.com/ohmpatel1997/findhotel/internal/service"
	"github.com/ohmpatel1997/findhotel/lib/config"
	"github.com/ohmpatel1997/findhotel/lib/db/init"
	"github.com/ohmpatel1997/findhotel/lib/log"
	"github.com/ohmpatel1997/findhotel/lib/router"
)

func main() {
	cfgPath := flag.String("p", "./cmd/client-api/config.yaml", "The configuration path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		panic(err)
	}

	l := zlog.New()
	l.Info("### Starting up client api ###", nil)

	host := os.Getenv("POSTGRES_HOST")
	dbName := os.Getenv("POSTGRES_DB")
	password := os.Getenv("POSTGRES_PASSWORD")
	user := os.Getenv("POSTGRES_USER")
	dbPort := os.Getenv("POSTGRES_PORT")
	conStr := fmt.Sprintf("postgres://%v:%v@%v:%v/%v", user, password, host, dbPort, dbName)

	db, err := pgsql.New(cfg.DB, conStr)
	if err != nil {
		panic(err)
	}

	manager := model.NewGeoLocationManager(db)

	srv := service.NewGeolocationService(manager)
	cntrl := controller.NewController(srv)
	router := registerRoutes(cntrl)

	err = router.ListenAndServeTLS(cfg.Server)
	if err != nil {
		panic(err)
	}
}

func registerRoutes(clientCntrl controller.ClientController) router.Router {
	r := router.NewBasicRouter()

	r.Route(clientCntrl.GetAPIVersionPath("/ip-info"), func(r router.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			clientCntrl.GetGeolocationData(w, r)
		})
	})

	return r
}
