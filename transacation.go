package go_orm

import (
	"context"
	"database/sql"
	"errors"
)

type session interface {
	getCore() core
	queryContext(context.Context, string, ...any) (*sql.Rows, error)
	execContext(context.Context, string, ...any) (sql.Result, error)
}

type Transaction struct {
	db *DB
	tx *sql.Tx
}

func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

func (t *Transaction) RollBack() error {
	return t.tx.Rollback()
}

func (t *Transaction) RollBackUnlessCommit() error {
	err := t.tx.Rollback()
	if !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}

func (t *Transaction) getCore() core {
	return t.db.core
}

func (t *Transaction) queryContext(ctx context.Context, s string, a ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, s, a...)
}

func (t *Transaction) execContext(ctx context.Context, s string, a ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, s, a...)
}
