package go_orm

import (
	"context"
)

type Selector[T any] struct {
	tableName   string
	where       []Predicate
	selectables []Selectable
	db          *DB
	builder     *builder
}

type Selectable interface {
	selectable()
}

func NewSelector[T any](db *DB) *Selector[T] {

	return &Selector[T]{
		db:    db,
		where: make([]Predicate, 0, 4),
	}
}

func (s *Selector[T]) Build(ctx context.Context) (*Query, error) {
	m, err := s.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	s.builder = NewBuilder(m, s.db.dialect)
	err = s.buildSelectables()
	if err != nil {
		return nil, err
	}

	if s.tableName == "" {
		s.builder.quote(s.builder.m.tableName)
	} else {
		s.builder.buildString(s.tableName)
	}
	if len(s.where) > 0 {
		s.builder.buildString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		err = s.builder.buildExpression(p)
		if err != nil {
			return nil, err
		}
	}
	s.builder.buildByte(';')
	return &Query{
		SQL:  s.builder.getSQL(),
		Args: s.builder.getArgs(),
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
	uac, err := NewUnsafeAccessor(s.builder.m, t)
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
		uac, err := NewUnsafeAccessor(s.builder.m, t)
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
		s.builder.buildString("SELECT ")

		for idx, selectable := range s.selectables {
			if idx > 0 {
				s.builder.buildString(", ")
			}
			switch se := selectable.(type) {
			case Column:
				if err := s.builder.buildColumn(se); err != nil {
					return err
				}
			case Aggregate:
				if err := s.builder.buildAggregate(se); err != nil {
					return err
				}
			case RawExpression:
				s.builder.buildString(se.expression)
				s.builder.addArgs(se.args...)
			}

		}
		s.builder.buildString(" FROM ")
	} else {
		s.builder.buildString("SELECT * FROM ")
	}
	return nil
}
