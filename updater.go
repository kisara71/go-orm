package go_orm

import (
	"context"
	"reflect"
)

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
func (u *Updater[T]) Build(ctx context.Context) (*Query, error) {

	m, err := u.core.registry.Get(new(T))
	if err != nil {
		return nil, err
	}

	u.builder = NewBuilder(m, u.core.dialect)

	u.builder.buildString("UPDATE ")
	u.builder.quote(u.builder.m.tableName)
	u.builder.buildString(" SET ")

	if u.val != nil {
		val := reflect.ValueOf(u.val).Elem()
		idx := 0
		for _, fd := range u.builder.m.fields {
			fieldVal := val.FieldByName(fd.goName)
			if fieldVal.IsZero() {
				continue
			}
			if idx > 0 {
				u.builder.buildString(", ")
			}
			u.builder.quote(fd.colName)
			u.builder.buildString(" = ?")
			u.builder.addArgs(fieldVal.Interface())
			idx++
		}
	} else {
		if len(u.assigns) == 0 {
			return nil, ErrUpdateNoColumns
		}
		for idx, assign := range u.assigns {
			if idx > 0 {
				u.builder.buildString(", ")
			}
			switch a := assign.(type) {
			case Assignment:
				if err := u.builder.buildColumn(a.column); err != nil {
					return nil, err
				}
				u.builder.buildString(" = ?")
				u.builder.addArgs(a.val)
			default:
				return nil, ErrUnsupportedType
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
	return &Query{
		SQL:  u.builder.getSQL(),
		Args: u.builder.getArgs(),
	}, nil
}
func (u *Updater[T]) Exec(ctx context.Context) *Result {
	query, err := u.Build(ctx)
	if err != nil {
		return &Result{
			err: err,
		}
	}
	res, err := u.sess.execContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return &Result{
			err: err,
		}
	}
	return &Result{
		res: res,
		err: nil,
	}
}
