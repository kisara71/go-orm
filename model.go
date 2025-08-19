package go_orm

import (
	"github.com/kisara71/go-orm/utils"
	"reflect"
	"strings"
	"sync"
)

type model struct {
	tableName string
	fields    map[string]*fieldInfo
	colMap    map[string]*fieldInfo
}
type TableName interface {
	TableName() string
}

var tableNameType = reflect.TypeOf((*TableName)(nil)).Elem()

type fieldInfo struct {
	colName string
	goName  string
	typ     reflect.Type
}
type registry struct {
	models sync.Map
}

const (
	columnTag = "column"
)

func (r *registry) Get(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, ErrInvalidModel
	}
	m, ok := r.models.Load(typ)
	if !ok {
		var err error
		m, err = r.parseModel(typ)
		if err != nil {
			return nil, err
		}
		r.models.Store(typ, m.(*model))
	}
	return m.(*model), nil
}
func (r *registry) parseModel(typ reflect.Type) (*model, error) {
	numField := typ.NumField()
	fields := make(map[string]*fieldInfo, numField)
	colMap := make(map[string]*fieldInfo, numField)
	for i := 0; i < numField; i++ {
		tags, err := r.parseTag(typ.Field(i).Tag)
		if err != nil {
			return nil, err
		}
		colName, ok := tags["column"]
		if !ok || colName == "" {
			colName = utils.CamelToSnake(typ.Field(i).Name)
		}
		fi := &fieldInfo{
			colName: colName,
			goName:  typ.Field(i).Name,
			typ:     typ.Field(i).Type,
		}
		fields[typ.Field(i).Name] = fi
		colMap[colName] = fi
	}
	var tableName string
	if reflect.PointerTo(typ).Implements(tableNameType) {
		tableName = reflect.New(typ).Interface().(TableName).TableName()
	} else {
		tableName = utils.CamelToSnake(typ.Name())
	}
	return &model{
		tableName: tableName,
		fields:    fields,
		colMap:    colMap,
	}, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	res := make(map[string]string, 4)
	fullTag, ok := tag.Lookup("orm")
	if !ok {
		return res, nil
	}
	pairs := strings.Split(strings.TrimSpace(fullTag), ",")
	for _, pair := range pairs {
		seg := strings.SplitN(pair, "=", 2)
		if len(seg) != 2 {
			return nil, ErrInvalidTags
		}
		res[seg[0]] = seg[1]
	}
	return res, nil
}
