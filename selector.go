package go_orm

import (
	"context"
)

type Selector[T any] struct {
	tableName   string
	where       []Predicate
	selectables []Selectable
	core        core
	sess        session
	builder     *builder
	groupExpr   []Expression
	having      []Predicate
	order       []OrderBy
	limit       int64
	offset      int64
}
type OrderBy struct {
	col   Column
	order string
}

func ASC(col string) OrderBy {
	return OrderBy{
		col:   C(col),
		order: "ASC",
	}
}
func DESC(col string) OrderBy {
	return OrderBy{
		col:   C(col),
		order: "DESC",
	}
}

type Selectable interface {
	selectable()
}

func NewSelector[T any](sess session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		core:      c,
		where:     make([]Predicate, 0, 4),
		groupExpr: make([]Expression, 0, 4),
		sess:      sess,
		order:     make([]OrderBy, 0, 2),
		having:    make([]Predicate, 0, 4),
	}
}

func (s *Selector[T]) Build(ctx context.Context) (*Query, error) {
	m, err := s.core.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	s.builder = NewBuilder(m, s.core.dialect)
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
		err = s.builder.buildExpression(p, ClauseWhere)
		if err != nil {
			return nil, err
		}
	}
	if len(s.groupExpr) > 0 {
		s.builder.buildString(" GROUP BY ")
		for idx, expr := range s.groupExpr {
			if idx > 0 {
				s.builder.buildString(", ")
			}
			switch exp := expr.(type) {
			case Column:
				if err := s.builder.buildColumn(exp); err != nil {
					return nil, err
				}
			case RawExpression:
				s.builder.buildString(exp.expression)
				s.builder.addArgs(exp.args...)
			default:
				return nil, ErrUnsupportedType
			}
		}
	}
	if len(s.having) > 0 {
		s.builder.buildString(" HAVING ")
		p := s.having[0]
		for i := 1; i < len(s.having); i++ {
			p = p.And(s.having[i])
		}
		err = s.builder.buildExpression(p, ClauseHaving)
		if err != nil {
			return nil, err
		}
	}
	if len(s.order) > 0 {
		s.builder.buildString(" ORDER BY ")
		for idx, order := range s.order {
			if idx > 0 {
				s.builder.buildString(", ")
			}
			if err := s.builder.buildColumn(order.col); err != nil {
				return nil, err
			}
			s.builder.buildByte(' ')
			s.builder.buildString(order.order)
		}
	}
	if s.limit > 0 {
		s.builder.buildString(" LIMIT ?")
		s.builder.addArgs(s.limit)
	}
	if s.offset > 0 {
		s.builder.buildString(" OFFSET ?")
		s.builder.addArgs(s.offset)
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

	rows, err := s.sess.queryContext(ctx, query.SQL, query.Args...)
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

func (s *Selector[T]) GroupBy(expr ...Expression) *Selector[T] {
	s.groupExpr = append(s.groupExpr, expr...)
	return s
}
func (s *Selector[T]) Having(p ...Predicate) *Selector[T] {
	s.having = append(s.having, p...)
	return s
}

func (s *Selector[T]) OrderBy(by ...OrderBy) *Selector[T] {
	s.order = append(s.order, by...)
	return s
}

func (s *Selector[T]) Limit(limit int64) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) Offset(offset int64) *Selector[T] {
	s.offset = offset
	return s
}
func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	query, err := s.Build(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.sess.queryContext(ctx, query.SQL, query.Args...)
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
