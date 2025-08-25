package trace

import (
	"github.com/kisara71/go-orm/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type MiddleWareBuilder struct {
	tracer trace.Tracer
}

func New(tracer trace.Tracer) MiddleWareBuilder {
	return MiddleWareBuilder{tracer: tracer}
}
func (m *MiddleWareBuilder) Build() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx *middleware.Context) *middleware.Result {
			var span trace.Span
			ctx.Ctx, span = m.tracer.Start(ctx.Ctx, "sql")
			res := next(ctx)
			span.SetAttributes(attribute.String("query:", ctx.Statement))
			span.End()
			return res
		}
	}
}
