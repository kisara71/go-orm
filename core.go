package go_orm

import "github.com/kisara71/go-orm/middleware"

type core struct {
	registry *registry
	dialect  Dialect
	mdls     []middleware.Middleware
}
