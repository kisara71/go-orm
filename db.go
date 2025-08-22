package go_orm

import "database/sql"

type DB struct {
	registry *registry
	db       *sql.DB
	dialect  Dialect
}
type DBOptions func(db *DB)

func Open(driver string, dsn string, options ...DBOptions) (*DB, error) {
	sqldb, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	db := &DB{
		registry: &registry{},
		db:       sqldb,
		dialect:  StandardSQL,
	}
	for _, opt := range options {
		opt(db)
	}
	return db, nil
}
func OpenDB(sqldb *sql.DB, options ...DBOptions) *DB {
	db := &DB{
		db:       sqldb,
		registry: &registry{},
		dialect:  StandardSQL,
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
