package go_orm

import (
	"context"
	"reflect"
	"strings"
)

type Insertor[T any] struct {
	args           []any
	columns        []string
	db             *DB
	m              *model
	sb             *strings.Builder
	values         []*T
	onDuplicateKey *OnDuplicateKey
}

func NewInsertor[T any](db *DB) *Insertor[T] {
	return &Insertor[T]{
		db: db,
		sb: &strings.Builder{},
	}
}

type OnDuplicateKeyBuilder[T any] struct {
	i *Insertor[T]
}

type OnDuplicateKey struct {
	assigns []Assignable
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Insertor[T] {
	o.i.onDuplicateKey = &OnDuplicateKey{
		assigns: assigns,
	}
	return o.i
}

func (i *Insertor[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

func (i *Insertor[T]) Build(ctx context.Context) (*Query, error) {
	if len(i.values) == 0 {
		return nil, ErrInsertNoValues
	}
	var err error
	i.m, err = i.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	i.sb.WriteString("INSERT INTO ")
	i.sb.WriteByte('`')
	i.sb.WriteString(i.m.tableName)
	i.sb.WriteString("` ")

	fields := i.m.fields
	if len(i.columns) == 0 {
		i.sb.WriteByte('(')
		for idx, fd := range i.m.fields {
			if idx > 0 {
				i.sb.WriteString(", ")
			}
			i.sb.WriteByte('`')
			i.sb.WriteString(fd.colName)
			i.sb.WriteByte('`')
		}
		i.sb.WriteByte(')')
	} else {
		fields = make([]*fieldInfo, 0, len(i.columns))
		i.sb.WriteByte('(')
		for idx, col := range i.columns {
			if fd, ok := i.m.goMap[col]; !ok {
				return nil, ErrUnknownField
			} else {
				fields = append(fields, fd)
			}
			if idx > 0 {
				i.sb.WriteString(", ")
			}
			i.sb.WriteByte('`')
			i.sb.WriteString(fields[idx].colName)
			i.sb.WriteByte('`')
		}
		i.sb.WriteByte(')')
	}

	i.sb.WriteString(" VALUES ")
	for idx1, val := range i.values {
		if idx1 > 0 {
			i.sb.WriteString(", ")
		}
		rval := reflect.ValueOf(val).Elem()
		i.sb.WriteByte('(')
		for idx2, fd := range fields {
			if idx2 > 0 {
				i.sb.WriteString(", ")
			}
			i.sb.WriteByte('?')
			i.args = append(i.args, rval.FieldByName(fd.goName).Interface())
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicateKey != nil {
		i.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		for idx, assign := range i.onDuplicateKey.assigns {
			if idx > 0 {
				i.sb.WriteString(", ")
			}
			switch as := assign.(type) {
			case Assignment:
				if err := buildColumns(C(as.column), i.sb, i.m.goMap); err != nil {
					return nil, err
				}
				i.sb.WriteString(" = ?")
				i.args = append(i.args, as.val)
			case Column:
				err := buildColumns(as, i.sb, i.m.goMap)
				if err != nil {
					return nil, ErrUnknownField
				}
				i.sb.WriteString(" = VALUES(")
				_ = buildColumns(as, i.sb, i.m.goMap)
				i.sb.WriteByte(')')
			}
		}
	}
	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

func (i *Insertor[T]) Values(vals ...*T) *Insertor[T] {
	i.values = vals
	return i
}

func (i *Insertor[T]) Columns(cols ...string) *Insertor[T] {
	i.columns = cols
	return i
}
