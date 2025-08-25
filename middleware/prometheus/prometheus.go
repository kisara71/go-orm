package prometheus

import (
	"github.com/kisara71/go-orm/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MiddleWareBuilder struct {
	summary *prometheus.SummaryVec
}

func New(namespace string, subsystem string) MiddleWareBuilder {
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "go_orm_prometheus",
		Help:      "record the sql exec duration",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.8:  0.02,
			0.9:  0.01,
			0.99: 0.001,
		},
	}, []string{"type", "table"})

	prometheus.MustRegister(summary)
	return MiddleWareBuilder{summary: summary}
}

func (m MiddleWareBuilder) Build() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx *middleware.Context) *middleware.Result {
			start := time.Now()
			res := next(ctx)
			m.summary.WithLabelValues(ctx.Type.String(), ctx.Model.TableName).Observe(time.Since(start).Seconds())
			return res
		}
	}
}
