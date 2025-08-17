package go_orm

import (
	"context"
	"strings"
)

type Selector[T any] struct {
	tableName string
	m         *model
	sb        *strings.Builder
	where     []Predicate
	args      []any
}

func (s *Selector[T]) Build(ctx context.Context) (*Query, error) {
	s.sb = &strings.Builder{}
	s.sb.WriteString("SELECT * FROM ")
	var err error
	s.m, err = parseModel(new(T))
	if err != nil {
		return nil, err
	}
	if s.tableName == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.m.tableName)
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
		err = buildExpression(s.sb, &s.args, p, s.m.fields)
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
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}
