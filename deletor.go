package go_orm

import (
	"database/sql"
	"github.com/kisara71/go-orm/middleware"
)

var _ Builder = &Deletor[any]{}

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

func (d *Deletor[T]) Build(ctx *middleware.Context) error {
	m, err := d.core.registry.Get(new(T))
	if err != nil {
		return err
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
			return err
		}
	}
	ctx.SetArgs(d.builder.getArgs())
	ctx.SetStatement(d.builder.getSQL())
	return nil
}
func (d *Deletor[T]) From(tableName string) {
	d.tableName = tableName
}
func (d *Deletor[T]) Where(predicate ...Predicate) {
	d.where = predicate
}

var _ middleware.Handler = (&Deletor[any]{}).handleExec

func (d *Deletor[T]) handleExec(ctx *middleware.Context) *middleware.Result {
	err := d.Build(ctx)
	if err != nil {
		return &middleware.Result{
			Res: nil,
			Err: err,
		}
	}
	res, err := d.sess.execContext(ctx.Ctx, ctx.Statement, ctx.Args...)
	if err != nil {
		return &middleware.Result{
			Res: nil,
			Err: err,
		}
	}
	return &middleware.Result{
		Res: &ExecResult{
			res: res,
			err: nil,
		},
		Err: nil,
	}
}
func (d *Deletor[T]) Exec(ctx *middleware.Context) *ExecResult {
	ctx.Type = middleware.OpExec
	root := d.handleExec
	for i := len(d.core.mdls) - 1; i >= 0; i-- {
		root = d.core.mdls[i](root)
	}
	res := root(ctx)
	if res.Err != nil {
		return &ExecResult{
			res: nil,
			err: res.Err,
		}
	}
	return &ExecResult{
		res: res.Res.(sql.Result),
		err: nil,
	}
}
