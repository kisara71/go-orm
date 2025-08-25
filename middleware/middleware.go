package middleware

import (
	"context"
	"github.com/kisara71/go-orm/model"
)

type Result struct {
	Res any
	Err error
}

type Middleware func(handler Handler) Handler

type Handler func(ctx *Context) *Result
type OpType uint8

const (
	OpQuery OpType = iota
	OpExec
)

func (o OpType) String() string {
	switch o {
	case OpQuery:
		return "query"
	case OpExec:
		return "Execute"
	}
	return ""
}

type Context struct {
	Ctx       context.Context
	Model     *model.Model
	Statement string
	Type      OpType
	Args      []any
}

func (c *Context) SetStatement(statement string) {
	c.Statement = statement
}
func (c *Context) SetArgs(args []any) {
	c.Args = args
}
