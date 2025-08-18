package go_orm

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeletor_Build(t *testing.T) {
	type TestModel struct {
		Name string
		Age  int
	}
	db := &DB{
		registry: &registry{},
	}
	testCases := []struct {
		name      string
		builder   *Deletor[TestModel]
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "delete all",
			builder: NewDeletor[TestModel](db),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model`",
				Args: []any{},
			},
			wantErr: nil,
		},
		{
			name:    "delete with table override",
			builder: func() *Deletor[TestModel] { d := NewDeletor[TestModel](db); d.From("`table_test`"); return d }(),
			wantQuery: &Query{
				SQL:  "DELETE FROM `table_test`",
				Args: []any{},
			},
			wantErr: nil,
		},
		{
			name: "delete with where single",
			builder: func() *Deletor[TestModel] {
				d := NewDeletor[TestModel](db)
				d.Where(C("Name").Eq("hha"))
				return d
			}(),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `name` = ?",
				Args: []any{"hha"},
			},
			wantErr: nil,
		},
		{
			name: "delete with where and",
			builder: func() *Deletor[TestModel] {
				d := NewDeletor[TestModel](db)
				d.Where(C("Age").Eq(111), C("Name").Eq("hha"))
				return d
			}(),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE (`age` = ?) AND (`name` = ?)",
				Args: []any{111, "hha"},
			},
			wantErr: nil,
		},
		{
			name: "delete with where not",
			builder: func() *Deletor[TestModel] {
				d := NewDeletor[TestModel](db)
				d.Where(Not(C("Age").Eq(111)))
				return d
			}(),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE NOT (`age` = ?)",
				Args: []any{111},
			},
			wantErr: nil,
		},
		{
			name: "delete with invalid field",
			builder: func() *Deletor[TestModel] {
				d := NewDeletor[TestModel](db)
				d.Where(Not(C("fsfs").Eq(111)))
				return d
			}(),
			wantQuery: nil,
			wantErr:   ErrInvalidField,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.builder.Build(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, res)
		})
	}
}
