package go_orm

import (
	"context"
)

type Builder interface {
	Build(ctx context.Context) (*Query, error)
}

type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

type Executor interface {
	Exec(ctx context.Context) *Result
}
type Query struct {
	SQL  string
	Args []any
}
