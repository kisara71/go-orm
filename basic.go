package go_orm

import (
	"context"
	"github.com/kisara71/go-orm/middleware"
)

type Builder interface {
	Build(ctx *middleware.Context) error
}

type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

type Executor interface {
	Exec(ctx context.Context) *ExecResult
}
type Query struct {
	SQL  string
	Args []any
}
