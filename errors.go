package go_orm

import "errors"

var (
	ErrInvalidModel     = errors.New("invalid type of model, expected struct or pointer")
	ErrUnknownField     = errors.New("field not exists")
	ErrInvalidTags      = errors.New("invalid tags, check structure's tags")
	ErrUnknownColumn    = errors.New("get unknown columns")
	ErrNoRecord         = errors.New("found no record in database")
	ErrScanFailed       = errors.New("scan data failed, may get unknown columns")
	ErrInvalidArguments = errors.New("invalid arguments")
	ErrInsertNoValues   = errors.New("call insert with out values")
)
