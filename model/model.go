package model

import (
	"github.com/kisara71/go-orm/errs"
	"github.com/kisara71/go-orm/utils"
	"reflect"
	"strings"
	"sync"
)

type Model struct {
	TableName string
	Fields    []*FieldInfo
	GoMap     map[string]*FieldInfo
	ColMap    map[string]*FieldInfo
}
type TableName interface {
	TableName() string
}

var tableNameType = reflect.TypeOf((*TableName)(nil)).Elem()

type FieldInfo struct {
	ColName string
	GoName  string
	Type    reflect.Type
	Offset  uintptr
}
type Registry struct {
	models sync.Map
}

const (
	columnTag = "column"
)

func (r *Registry) Get(entity any) (*Model, error) {
	typ := reflect.TypeOf(entity)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errs.ErrInvalidModel
	}
	m, ok := r.models.Load(typ)
	if !ok {
		var err error
		m, err = r.parseModel(typ)
		if err != nil {
			return nil, err
		}
		r.models.Store(typ, m.(*Model))
	}
	return m.(*Model), nil
}
func (r *Registry) parseModel(typ reflect.Type) (*Model, error) {
	numField := typ.NumField()
	fields := make([]*FieldInfo, 0, numField)
	goMap := make(map[string]*FieldInfo, numField)
	colMap := make(map[string]*FieldInfo, numField)
	for i := 0; i < numField; i++ {
		tags, err := r.parseTag(typ.Field(i).Tag)
		if err != nil {
			return nil, err
		}
		colName, ok := tags["column"]
		if !ok || colName == "" {
			colName = utils.CamelToSnake(typ.Field(i).Name)
		}
		fi := &FieldInfo{
			ColName: colName,
			GoName:  typ.Field(i).Name,
			Type:    typ.Field(i).Type,
			Offset:  typ.Field(i).Offset,
		}
		goMap[typ.Field(i).Name] = fi
		colMap[colName] = fi
		fields = append(fields, fi)
	}
	var tableName string
	if reflect.PointerTo(typ).Implements(tableNameType) {
		tableName = reflect.New(typ).Interface().(TableName).TableName()
	} else {
		tableName = utils.CamelToSnake(typ.Name())
	}
	return &Model{
		TableName: tableName,
		GoMap:     goMap,
		ColMap:    colMap,
		Fields:    fields,
	}, nil
}

func (r *Registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	res := make(map[string]string, 4)
	fullTag, ok := tag.Lookup("orm")
	if !ok {
		return res, nil
	}
	pairs := strings.Split(strings.TrimSpace(fullTag), ",")
	for _, pair := range pairs {
		seg := strings.SplitN(pair, "=", 2)
		if len(seg) != 2 {
			return nil, errs.ErrInvalidTags
		}
		res[seg[0]] = seg[1]
	}
	return res, nil
}
