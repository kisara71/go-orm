package middleware

import "context"

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

type Context struct {
	Ctx       context.Context
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
