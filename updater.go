package go_orm

import (
	"github.com/kisara71/go-orm/errs"
	"github.com/kisara71/go-orm/middleware"
	"reflect"
)

var _ Builder = &Updater[any]{}

type Updater[T any] struct {
	builder *builder
	val     *T
	assigns []Assignable
	where   []Predicate
	core    core
	sess    session
}

func NewUpdater[T any](sess session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		core:    c,
		assigns: make([]Assignable, 0, 8),
		where:   make([]Predicate, 0, 4),
		sess:    sess,
	}
}

func (u *Updater[T]) Set(assign ...Assignable) *Updater[T] {
	u.assigns = append(u.assigns, assign...)
	return u
}

func (u *Updater[T]) FromStruct(val *T) *Updater[T] {
	u.val = val
	return u
}

func (u *Updater[T]) Where(p ...Predicate) *Updater[T] {
	u.where = append(u.where, p...)
	return u
}
func (u *Updater[T]) Build(ctx *middleware.Context) error {

	m, err := u.core.registry.Get(new(T))
	if err != nil {
		return err
	}

	u.builder = NewBuilder(m, u.core.dialect)

	u.builder.buildString("UPDATE ")
	u.builder.quote(u.builder.m.TableName)
	u.builder.buildString(" SET ")

	if u.val != nil {
		val := reflect.ValueOf(u.val).Elem()
		idx := 0
		for _, fd := range u.builder.m.Fields {
			fieldVal := val.FieldByName(fd.GoName)
			if fieldVal.IsZero() {
				continue
			}
			if idx > 0 {
				u.builder.buildString(", ")
			}
			u.builder.quote(fd.ColName)
			u.builder.buildString(" = ?")
			u.builder.addArgs(fieldVal.Interface())
			idx++
		}
	} else {
		if len(u.assigns) == 0 {
			return errs.ErrUpdateNoColumns
		}
		for idx, assign := range u.assigns {
			if idx > 0 {
				u.builder.buildString(", ")
			}
			switch a := assign.(type) {
			case Assignment:
				if err := u.builder.buildColumn(a.column); err != nil {
					return err
				}
				u.builder.buildString(" = ?")
				u.builder.addArgs(a.val)
			default:
				return errs.ErrUnsupportedType
			}
		}
	}
	if len(u.where) > 0 {
		u.builder.buildString(" WHERE ")
		p := u.where[0]
		for i := 1; i < len(u.where); i++ {
			p = p.And(u.where[i])
		}
		err = u.builder.buildExpression(p, ClauseWhere)
	}
	u.builder.buildByte(';')
	ctx.SetStatement(u.builder.getSQL())
	ctx.SetArgs(u.builder.getArgs())
	return nil
}

var _ middleware.Handler = (&Updater[any]{}).handleExec

func (u *Updater[T]) handleExec(ctx *middleware.Context) *middleware.Result {
	err := u.Build(ctx)
	if err != nil {
		return &middleware.Result{
			Res: nil,
			Err: err,
		}
	}
	res, err := u.sess.execContext(ctx.Ctx, ctx.Statement, ctx.Args...)
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
func (u *Updater[T]) Exec(ctx *middleware.Context) *ExecResult {
	ctx.Type = middleware.OpExec
	root := u.handleExec
	for i := len(u.core.mdls) - 1; i >= 0; i-- {
		root = u.core.mdls[i](root)
	}
	res := root(ctx)
	if res.Err != nil {
		return &ExecResult{
			res: nil,
			err: res.Err,
		}
	}
	return res.Res.(*ExecResult)
}
