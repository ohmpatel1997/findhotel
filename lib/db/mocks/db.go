package mocks

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/joho/godotenv"
	"github.com/ohmpatel1997/findhotel/lib/config"
	pgsql "github.com/ohmpatel1997/findhotel/lib/db/init"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
)

// NewPGContainer instantiates new PostgreSQL docker container
func NewPGContainer(t *testing.T) (*dockertest.Pool, *dockertest.Resource) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	pool.MaxWait = DockerTimeout(t)

	runOpts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=imploy",
		},
		Auth: *DockerHubAuth(t),
	}

	resource, err := pool.RunWithOptions(&runOpts)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	return pool, resource
}

func DockerHubAuth(t *testing.T) *dc.AuthConfiguration {
	err := loadEnvVars(t)
	if err != nil {
		t.Errorf("error loading environment: %s", err.Error())
	}
	res := &dc.AuthConfiguration{
		Username: os.Getenv("DOCKER_USERNAME"),
		Password: os.Getenv("DOCKER_ACCESS_TOKEN"),
	}
	return res
}

// DockerTimeout returns MaxWait duration for dockertest download
func DockerTimeout(t *testing.T) time.Duration {
	err := loadEnvVars(t)
	if err != nil {
		t.Errorf("error loading environment: %s", err.Error())
	}

	var maxTime int64
	maxTime, err = strconv.ParseInt(os.Getenv("DOCKER_DOWNLOAD_TIMEOUT_MIN"), 10, 64)
	if err != nil {
		t.Errorf("error parsing Docker download timeout: %s", err.Error())
		maxTime = 5
	}
	return time.Minute * time.Duration(maxTime)
}

func loadEnvVars(t *testing.T) error {
	env := os.Getenv("ENV")

	if len(env) == 0 {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return errors.New("caller information not found")
		}
		filePath := path.Dir(filename) + "/../../../.env.test"
		t.Logf("test envs loaded from file: %s", filePath)
		return godotenv.Load(filePath)
	}
	return nil
}

// NewDBArray instantiates new postgresql database connection via docker container
func NewDBArray(t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource, models []interface{}) *pg.DB {
	var db *pg.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = pgsql.New(
			&config.Database{
				Timeout: 10,
				SSLMode: false,
			}, fmt.Sprintf("postgres://postgres:secret@localhost:%s/%s", resource.GetPort("5432/tcp"), "imploy"))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	for _, v := range models {
		t.Fatal(db.Model(v).CreateTable(&orm.CreateTableOptions{FKConstraints: true}))
	}

	return db
}

// NewDB instantiates new postgresql database connection via docker container
func NewDB(t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource, models ...interface{}) *pg.DB {
	return NewDBArray(t, pool, resource, models)
}

// CloseContainer closes docker container
func CloseContainer(t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		t.Fatalf("Could not purge resource: %s", err)
	}
}
