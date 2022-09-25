// Package db holds the DB interface
package db

import (
	"context"
	"io"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// DB represents all the methods required for pg-db
type DB interface {
	Begin() (*pg.Tx, error)
	RunInTransaction(ctx context.Context, fn func(*pg.Tx) error) error
	Model(model ...interface{}) *orm.Query
	ModelContext(c context.Context, model ...interface{}) *orm.Query

	Exec(query interface{}, params ...interface{}) (pg.Result, error)
	ExecContext(c context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	ExecOne(query interface{}, params ...interface{}) (pg.Result, error)
	ExecOneContext(c context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	Query(model, query interface{}, params ...interface{}) (pg.Result, error)
	QueryContext(c context.Context, model, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOne(model, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOneContext(c context.Context, model, query interface{}, params ...interface{}) (pg.Result, error)

	CopyFrom(r io.Reader, query interface{}, params ...interface{}) (pg.Result, error)
	CopyTo(w io.Writer, query interface{}, params ...interface{}) (pg.Result, error)

	Context() context.Context
}
