package go_orm

import (
	"context"
	"database/sql"
	"github.com/kisara71/go-orm/middleware"
)

type DB struct {
	core
	db *sql.DB
}

func (d *DB) getCore() core {
	return d.core
}

func (d *DB) queryContext(ctx context.Context, s string, a ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, s, a...)
}

func (d *DB) execContext(ctx context.Context, s string, a ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, s, a...)
}

func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Transaction, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		db: d,
		tx: tx,
	}, nil
}
func (d *DB) DoTx(ctx context.Context, fn func(ctx context.Context, tx *Transaction) error) (err error) {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	tranz := &Transaction{
		db: d,
		tx: tx,
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			err = tranz.RollBack()
		} else {
			err = tranz.Commit()
		}
	}()
	err = fn(ctx, tranz)
	panicked = false

	return err
}
func (d *DB) Use(middlewares ...middleware.Middleware) {
	d.mdls = append(d.mdls, middlewares...)
}

type DBOptions func(db *DB)

func Open(driver string, dsn string, options ...DBOptions) (*DB, error) {
	sqldb, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	db := &DB{
		db: sqldb,
		core: core{
			registry: &registry{},
			dialect:  StandardSQL,
		},
	}
	for _, opt := range options {
		opt(db)
	}
	return db, nil
}
func OpenDB(sqldb *sql.DB, options ...DBOptions) *DB {
	db := &DB{
		db: sqldb,
		core: core{
			registry: &registry{},
			dialect:  StandardSQL,
		},
	}
	for _, opt := range options {
		opt(db)
	}
	return db
}

func WithDialect(dialect Dialect) DBOptions {
	return func(db *DB) {
		db.dialect = dialect
	}
}
