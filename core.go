package go_orm

import (
	"github.com/kisara71/go-orm/middleware"
	"github.com/kisara71/go-orm/model"
)

type core struct {
	registry *model.Registry
	dialect  Dialect
	mdls     []middleware.Middleware
}
