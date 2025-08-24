package log

import (
	"github.com/kisara71/go-orm/middleware"
	"log"
)

type MiddleWareBuilder struct {
	LogFunc func(query string, args []any)
}

func NewDefault() MiddleWareBuilder {
	return MiddleWareBuilder{
		LogFunc: func(query string, args []any) {
			log.Printf(`sql : "%s", args: %v`, query, args)
		},
	}
}
func New(logFunc func(query string, args []any)) MiddleWareBuilder {
	return MiddleWareBuilder{
		LogFunc: logFunc,
	}
}

func (m MiddleWareBuilder) Build() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx *middleware.Context) *middleware.Result {
			res := next(ctx)
			m.LogFunc(ctx.Statement, ctx.Args)
			return res
		}
	}
}
