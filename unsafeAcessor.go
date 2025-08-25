package go_orm

import (
	"database/sql"
	"errors"
	errors2 "github.com/kisara71/go-orm/errs"
	"github.com/kisara71/go-orm/model"
	"reflect"
	"unsafe"
)

type UnsafeAccessor interface {
	Set(rows *sql.Rows) error
	Fetch(field string) (any, error)
	Access(entity any)
}

type unsafeAccessor struct {
	m      *model.Model
	entity any
}

func NewUnsafeAccessor(model *model.Model) UnsafeAccessor {
	return &unsafeAccessor{
		m: model,
	}
}
func (u *unsafeAccessor) Set(rows *sql.Rows) error {
	if u.entity == nil {
		return errors.New("unsafe accessor has no vals")
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	address := reflect.ValueOf(u.entity).UnsafePointer()
	vals := make([]any, 0, len(u.m.GoMap))
	for _, col := range cols {
		fd, ok := u.m.ColMap[col]
		if !ok {
			return errors2.ErrUnknownColumn
		}
		vals = append(vals,
			reflect.NewAt(fd.Type, unsafe.Pointer((uintptr)(address)+fd.Offset)).Interface())
	}
	err = rows.Scan(vals...)
	if err != nil {
		return errors2.ErrScanFailed
	}
	return nil
}

func (u *unsafeAccessor) Access(entity any) {
	u.entity = entity
}

func (u *unsafeAccessor) Fetch(field string) (any, error) {
	if fd, ok := u.m.GoMap[field]; !ok {
		return nil, errors2.ErrUnknownField
	} else {
		address := reflect.ValueOf(u.entity).UnsafePointer()
		return reflect.NewAt(fd.Type,
			unsafe.Pointer((uintptr)(address)+fd.Offset)).Elem().Interface(), nil
	}
}
