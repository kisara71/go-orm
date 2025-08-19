package go_orm

import "database/sql"

type DB struct {
	registry *registry
	db       *sql.DB
}

func Open(driver string, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return &DB{
		registry: &registry{},
		db:       db,
	}, nil
}
func OpenDB(db *sql.DB) *DB {
	return &DB{
		db:       db,
		registry: &registry{},
	}
}
