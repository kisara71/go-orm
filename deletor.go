package go_orm

import (
	"context"
)

type Deletor[T any] struct {
	tableName string
	where     []Predicate
	db        *DB
	builder   *builder
}

func NewDeletor[T any](db *DB) *Deletor[T] {
	return &Deletor[T]{
		db:    db,
		where: make([]Predicate, 0, 4),
	}
}

func (d *Deletor[T]) Build(ctx context.Context) (*Query, error) {
	m, err := d.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	d.builder = NewBuilder(m, d.db.dialect)
	d.builder.buildString("DELETE FROM ")
	if d.tableName == "" {
		d.builder.quote(d.builder.m.tableName)
	} else {
		d.builder.buildString(d.tableName)
	}
	if len(d.where) > 0 {
		d.builder.buildString(" WHERE ")
		p := d.where[0]
		for i := 1; i < len(d.where); i++ {
			p = p.And(d.where[i])
		}
		err = d.builder.buildExpression(p, ClauseDelete)
		if err != nil {
			return nil, err
		}
	}
	return &Query{
		SQL:  d.builder.getSQL(),
		Args: d.builder.getArgs(),
	}, nil
}
func (d *Deletor[T]) From(tableName string) {
	d.tableName = tableName
}
func (d *Deletor[T]) Where(predicate ...Predicate) {
	d.where = predicate
}

func (d *Deletor[T]) Exec(ctx context.Context) *Result {
	query, err := d.Build(ctx)
	if err != nil {
		return &Result{
			err: err,
		}
	}
	res, err := d.db.db.ExecContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return &Result{
			err: err,
		}
	}
	return &Result{
		err: nil,
		res: res,
	}
}
