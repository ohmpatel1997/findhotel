package pgsql_test

import (
	"fmt"
	"imploy/lib/mock"
	"testing"

	"imploy/lib/config"
	pgsql "imploy/lib/pgsql/init"
	"imploy/lib/user"

	"github.com/go-pg/pg/v10/orm"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	pool, resource := mock.NewPGContainer(t)

	_, err := pgsql.New(&config.Database{
		LogQueries: false,
		Timeout:    1,
		SSLMode:    false,
	}, "PSN")
	if err == nil {
		t.Error("Expected error")
	}

	_, err = pgsql.New(
		&config.Database{
			LogQueries: false,
			Timeout:    0,
			SSLMode:    false,
		}, fmt.Sprintf("postgres://postgres:secret@localhost:%s/%s", resource.GetPort("1234/tcp"), "imploy"))
	if err == nil {
		t.Error("Expected error")
	}

	db := mock.NewDB(t, pool, resource)

	var usr user.User

	err = db.Model(&usr).CreateTable(&orm.CreateTableOptions{FKConstraints: true})
	if err != nil {
		t.Fatalf("Error creating schema %v", err)
	}

	usr = user.User{Email: "john@wick.com"}
	_, err = db.Model(&usr).Insert()
	if err != nil {
		t.Fatalf("Error inserting new user %v", err)
	}

	err = db.Model(&usr).WherePK().Select()
	if err != nil {
		t.Fatal("Error getting Users")
	}

	assert.NotNil(t, db)

	err = db.Close()
	if err != nil {
		t.Fatal("Error closing DB")
	}

	if err := pool.Purge(resource); err != nil {
		t.Fatalf("Could not purge resource: %s", err)
	}
}
