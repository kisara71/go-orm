package go_orm

import (
	"database/sql"
	"reflect"
	"unsafe"
)

type UnsafeAccessor interface {
	Set(rows *sql.Rows) error
	Fetch(field string) (any, error)
}

type unsafeAccessor struct {
	m      *model
	entity any
}

func NewUnsafeAccessor(model *model, entity any) (UnsafeAccessor, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidArguments
	}
	return &unsafeAccessor{
		m:      model,
		entity: entity,
	}, nil
}
func (u *unsafeAccessor) Set(rows *sql.Rows) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	address := reflect.ValueOf(u.entity).UnsafePointer()
	vals := make([]any, 0, len(u.m.goMap))
	for _, col := range cols {
		fd, ok := u.m.colMap[col]
		if !ok {
			return ErrUnknownColumn
		}
		vals = append(vals,
			reflect.NewAt(fd.typ, unsafe.Pointer((uintptr)(address)+fd.offset)).Interface())
	}
	err = rows.Scan(vals...)
	if err != nil {
		return ErrScanFailed
	}
	return nil
}

func (u *unsafeAccessor) Fetch(field string) (any, error) {
	if fd, ok := u.m.goMap[field]; !ok {
		return nil, ErrUnknownField
	} else {
		address := reflect.ValueOf(u.entity).UnsafePointer()
		return reflect.NewAt(fd.typ,
			unsafe.Pointer((uintptr)(address)+fd.offset)).Interface(), nil
	}
}
