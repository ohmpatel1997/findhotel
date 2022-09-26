package pgsql_test

import (
	"fmt"
	"testing"

	"github.com/ohmpatel1997/findhotel/internal/model"
	"github.com/ohmpatel1997/findhotel/lib/config"
	pgsql "github.com/ohmpatel1997/findhotel/lib/db/init"
	"github.com/ohmpatel1997/findhotel/lib/db/mocks"

	"github.com/go-pg/pg/v10/orm"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	pool, resource := mocks.NewPGContainer(t)
	defer mocks.CloseContainer(t, pool, resource)

	_, err := pgsql.New(&config.Database{
		Timeout: 1,
		SSLMode: false,
	}, "PSN")
	if err == nil {
		t.Error("Expected error")
	}

	_, err = pgsql.New(
		&config.Database{
			Timeout: 0,
			SSLMode: false,
		}, fmt.Sprintf("postgres://postgres:secret@localhost:%s/%s", resource.GetPort("1234/tcp"), "imploy"))
	if err == nil {
		t.Error("Expected error")
	}

	db := mocks.NewDB(t, pool, resource)

	var geo model.Geolocation

	err = db.Model(&geo).CreateTable(&orm.CreateTableOptions{FKConstraints: true})
	if err != nil {
		t.Fatalf("Error creating schema %v", err)
	}

	_, err = db.Model(&geo).Insert()
	if err != nil {
		t.Fatalf("Error inserting new user %v", err)
	}

	err = db.Model(&geo).WherePK().Select()
	if err != nil {
		t.Fatal("Error getting Users")
	}

	assert.NotNil(t, db)

	err = db.Close()
	if err != nil {
		t.Fatal("Error closing DB")
	}

}
