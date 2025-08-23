package go_orm

import (
	"context"
)

type Deletor[T any] struct {
	tableName string
	where     []Predicate
	builder   *builder
	sess      session
	core      core
}

func NewDeletor[T any](sess session) *Deletor[T] {
	c := sess.getCore()
	return &Deletor[T]{
		core:  c,
		where: make([]Predicate, 0, 4),
		sess:  sess,
	}
}

func (d *Deletor[T]) Build(ctx context.Context) (*Query, error) {
	m, err := d.core.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	d.builder = NewBuilder(m, d.core.dialect)
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
		err = d.builder.buildExpression(p, ClauseWhere)
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
	res, err := d.sess.execContext(ctx, query.SQL, query.Args...)
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
