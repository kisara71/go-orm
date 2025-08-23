package go_orm

import (
	"context"
)

type Insertor[T any] struct {
	columns    []string
	values     []*T
	onConflict *OnConflict
	core       core
	sess       session
	builder    *builder
}

func NewInsertor[T any](sess session) *Insertor[T] {
	c := sess.getCore()
	return &Insertor[T]{
		sess:   sess,
		core:   c,
		values: make([]*T, 0, 16),
	}
}

type OnConflictBuilder[T any] struct {
	i          *Insertor[T]
	onConflict *OnConflict
}

type OnConflict struct {
	assigns         []Assignable
	conflictColumns []Column
}

func (o *OnConflictBuilder[T]) Update(assigns ...Assignable) *Insertor[T] {
	o.onConflict.assigns = append(o.onConflict.assigns, assigns...)
	o.i.onConflict = o.onConflict
	return o.i
}
func (o *OnConflictBuilder[T]) Columns(cols ...Column) *OnConflictBuilder[T] {
	o.onConflict.conflictColumns = append(o.onConflict.conflictColumns, cols...)
	return o
}

func (i *Insertor[T]) OnConflict() *OnConflictBuilder[T] {
	return &OnConflictBuilder[T]{
		i: i,
		onConflict: &OnConflict{
			assigns:         make([]Assignable, 0, 8),
			conflictColumns: make([]Column, 0, 4),
		},
	}
}

func (i *Insertor[T]) Build(ctx context.Context) (*Query, error) {
	if len(i.values) == 0 {
		return nil, ErrInsertNoValues
	}
	m, err := i.core.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	i.builder = NewBuilder(m, i.core.dialect)
	i.builder.sb.WriteString("INSERT INTO ")

	i.builder.quote(i.builder.m.tableName)
	i.builder.buildByte(' ')
	fields := i.builder.m.fields
	if len(i.columns) == 0 {
		i.builder.buildByte('(')
		for idx, fd := range i.builder.m.fields {
			if idx > 0 {
				i.builder.sb.WriteString(", ")
			}
			i.builder.quote(fd.colName)
		}
		i.builder.buildByte(')')
	} else {
		fields = make([]*fieldInfo, 0, len(i.columns))
		i.builder.buildByte('(')
		for idx, col := range i.columns {
			if fd, ok := i.builder.m.goMap[col]; !ok {
				return nil, ErrUnknownField
			} else {
				fields = append(fields, fd)
			}
			if idx > 0 {
				i.builder.buildString(", ")
			}
			i.builder.quote(fields[idx].colName)
		}
		i.builder.buildByte(')')
	}

	i.builder.buildString(" VALUES ")
	for idx1, val := range i.values {
		if idx1 > 0 {
			i.builder.buildString(", ")
		}
		//rval := reflect.ValueOf(val).Elem()
		accessor, err := NewUnsafeAccessor(i.builder.m, val)
		if err != nil {
			return nil, err
		}
		i.builder.buildByte('(')
		for idx2, fd := range fields {
			if idx2 > 0 {
				i.builder.buildString(", ")
			}
			i.builder.buildByte('?')
			arg, err := accessor.Fetch(fd.goName)
			if err != nil {
				return nil, err
			}
			i.builder.addArgs(arg)
		}
		i.builder.buildByte(')')
	}

	if i.onConflict != nil {
		if err = i.core.dialect.BuildUpsert(i.builder, i.onConflict); err != nil {
			return nil, err
		}
	}
	i.builder.buildByte(';')
	return &Query{
		SQL:  i.builder.getSQL(),
		Args: i.builder.getArgs(),
	}, nil
}

func (i *Insertor[T]) Values(vals ...*T) *Insertor[T] {
	i.values = append(i.values, vals...)
	return i
}

func (i *Insertor[T]) Columns(cols ...string) *Insertor[T] {
	i.columns = cols
	return i
}
func (i *Insertor[T]) Exec(ctx context.Context) *Result {
	query, err := i.Build(ctx)
	if err != nil {
		return &Result{
			err: err,
		}
	}
	res, err := i.sess.execContext(ctx, query.SQL, query.Args...)
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
