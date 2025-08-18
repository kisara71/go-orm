package go_orm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 一个用来测试的 struct

// 用来测试 TableName 接口
type CustomTable struct {
	Name string
}

func (c *CustomTable) TableName() string {
	return "my_custom_table"
}

func TestRegistry_ParseModel(t *testing.T) {
	type TestModel struct {
		Id   int    `orm:"column=id_t"`
		Name string // 没有 tag，用 CamelToSnake
		Age  int    `orm:"column=age_t"`
	}
	r := &registry{}

	testCases := []struct {
		name    string
		entity  any
		wantTbl string
		wantCol map[string]string
		wantErr error
	}{
		{
			name:    "basic struct",
			entity:  TestModel{},
			wantTbl: "test_model",
			wantCol: map[string]string{
				"Id":   "id_t",
				"Name": "name",
				"Age":  "age_t",
			},
			wantErr: nil,
		},
		{
			name:    "ptr struct",
			entity:  &TestModel{},
			wantTbl: "test_model",
			wantCol: map[string]string{
				"Id":   "id_t",
				"Name": "name",
				"Age":  "age_t",
			},
			wantErr: nil,
		},
		{
			name:    "custom table name",
			entity:  &CustomTable{},
			wantTbl: "my_custom_table",
			wantCol: map[string]string{
				"Name": "name",
			},
			wantErr: nil,
		},
		{
			name:    "invalid type",
			entity:  123,
			wantErr: ErrInvalidModel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantTbl, m.tableName)
			gotCols := make(map[string]string, len(m.fields))
			for k, f := range m.fields {
				gotCols[k] = f.colName
			}
			assert.True(t, reflect.DeepEqual(tc.wantCol, gotCols))
		})
	}
}
