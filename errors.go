package go_orm

import "errors"

var (
	ErrInvalidModel = errors.New("invalid type of model, expected struct or pointer")
	ErrInvalidField = errors.New("field not exists")
	ErrInvalidTags  = errors.New("invalid tags, check structure's tags")
)
