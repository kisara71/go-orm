package go_orm

import (
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	tableName string
	sb        *strings.Builder
	where     []Predicate
	args      []any
	db        *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db:    db,
		sb:    &strings.Builder{},
		args:  make([]any, 0, 4),
		where: make([]Predicate, 0, 4),
	}
}

func (s *Selector[T]) Build(ctx context.Context) (*Query, error) {
	m, err := s.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT * FROM ")
	if s.tableName == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(m.tableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.tableName)
	}
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		s.args = make([]any, 0, 4)
		err = buildExpression(s.sb, &s.args, p, m.fields)
		if err != nil {
			return nil, err
		}
	}
	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.tableName = table
	return s
}

func (s *Selector[T]) Where(p ...Predicate) *Selector[T] {
	s.where = p
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Build(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.db.QueryContext(ctx, query.SQL, query.Args...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	t := new(T)
	m, err := s.db.registry.Get(t)
	if err != nil {
		return nil, err
	}
	vals := make([]any, 0, len(cols))
	for _, col := range cols {
		if fd, ok := m.colMap[col]; ok {
			vals = append(vals, reflect.New(fd.typ).Interface())
		} else {
			return nil, ErrUnknownColumn
		}
	}
	if !rows.Next() {
		return nil, ErrNoRecord
	}
	err = rows.Scan(vals...)
	if err != nil {
		return nil, ErrScanFailed
	}
	for idx, col := range cols {
		reflect.ValueOf(t).Elem().FieldByName(m.colMap[col].goName).
			Set(reflect.ValueOf(vals[idx]).Elem())
	}
	return t, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}
