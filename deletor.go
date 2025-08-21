package go_orm

import (
	"context"
	"strings"
)

type Deletor[T any] struct {
	sb        *strings.Builder
	tableName string
	where     []Predicate
	args      []any
	db        *DB
	m         *model
}

func NewDeletor[T any](db *DB) *Deletor[T] {
	return &Deletor[T]{
		db:    db,
		where: make([]Predicate, 0, 4),
		args:  make([]any, 0, 4),
		sb:    &strings.Builder{},
	}
}

func (d *Deletor[T]) Build(ctx context.Context) (*Query, error) {
	m, err := d.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	d.m = m
	d.sb = &strings.Builder{}
	d.sb.WriteString("DELETE FROM ")
	if d.tableName == "" {
		d.sb.WriteByte('`')
		d.sb.WriteString(m.tableName)
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.tableName)
	}
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		p := d.where[0]
		for i := 1; i < len(d.where); i++ {
			p = p.And(d.where[i])
		}
		d.args = make([]any, 0, 4)
		err = buildExpression(d.sb, &d.args, p, d.m.goMap)
		if err != nil {
			return nil, err
		}
	}
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}
func (d *Deletor[T]) From(tableName string) {
	d.tableName = tableName
}
func (d *Deletor[T]) Where(predicate ...Predicate) {
	d.where = predicate
}
