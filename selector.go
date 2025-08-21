package go_orm

import (
	"context"
	"strings"
)

type Selector[T any] struct {
	tableName   string
	sb          *strings.Builder
	where       []Predicate
	selectables []Selectable
	args        []any
	db          *DB
	m           *model
}

type Selectable interface {
	selectable()
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
	s.m = m
	err = s.buildSelectables()
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

func (s *Selector[T]) Select(selectables ...Selectable) *Selector[T] {
	s.selectables = selectables
	return s
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
	if !rows.Next() {
		return nil, ErrNoRecord
	}

	t := new(T)
	uac, err := NewUnsafeAccessor(s.m, t)
	if err != nil {
		return nil, err
	}
	err = uac.Set(rows)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	query, err := s.Build(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.db.QueryContext(ctx, query.SQL, query.Args...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	res := make([]*T, 0, 32)
	for rows.Next() {
		t := new(T)
		uac, err := NewUnsafeAccessor(s.m, t)
		if err != nil {
			return nil, err
		}
		err = uac.Set(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (s *Selector[T]) buildSelectables() error {
	if len(s.selectables) > 0 {
		s.sb.WriteString("SELECT ")

		for idx, selectable := range s.selectables {
			if idx > 0 {
				s.sb.WriteString(", ")
			}
			switch se := selectable.(type) {
			case Column:
				if err := buildColumns(se, s.sb, s.m.fields); err != nil {
					return err
				}
			case Aggregate:
				if err := buildAggregates(se, s.sb, s.m.fields); err != nil {
					return err
				}
			case RawExpression:
				s.sb.WriteString(se.expression)
				s.args = append(s.args, se.args...)
			}

		}
		s.sb.WriteString(" FROM ")
	} else {
		s.sb.WriteString("SELECT * FROM ")
	}
	return nil
}
