package go_orm

import (
	"github.com/kisara71/go-orm/utils"
	"reflect"
)

type model struct {
	tableName string
	fields    map[string]fieldInfo
}

type fieldInfo struct {
	colName string
}

func parseModel(m any) (*model, error) {
	typ := reflect.TypeOf(m)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, ErrInvalidModel
	}
	numField := typ.NumField()
	fields := make(map[string]fieldInfo, numField)
	for i := 0; i < numField; i++ {
		fields[typ.Field(i).Name] = fieldInfo{
			colName: utils.CamelToSnake(typ.Field(i).Name),
		}
	}
	return &model{
		tableName: utils.CamelToSnake(typ.Name()),
		fields:    fields,
	}, nil
}
