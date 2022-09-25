package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ohmpatel1997/findhotel/internal/model"
	"github.com/ohmpatel1997/findhotel/internal/service"
	"github.com/ohmpatel1997/findhotel/lib/config"
	pgsql "github.com/ohmpatel1997/findhotel/lib/db/init"
	zlog "github.com/ohmpatel1997/findhotel/lib/log"
)

func main() {
	_ = zlog.New()

	cfgPath := flag.String("p", "./cmd/import/config.yaml", "The configuration path")
	dumpFilePath := flag.String("s", "./cmd/import/data_dump.csv", "The configuration path")
	flag.Parse()
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		panic(err)
	}

	file, err := os.Open(*dumpFilePath)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

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
	parserService := service.NewParser(file, manager)

	timeTaken, invalid, validData, err := parserService.ParseAndStore()
	if err != nil {
		zlog.Logger().Error("error parsing", err, nil)
		return
	}

	zlog.Logger().Info("Successfully Parsed And Stored", nil)

	zlog.Logger().Info("Metrics: ", zlog.ParamsType{"Time Taken": timeTaken, "Valid Data": validData, "Invalid Data": invalid})
}
