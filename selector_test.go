package go_orm

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector(t *testing.T) {
	type TestModel struct {
		Name string
		Age  int
	}
	db := &DB{
		registry: &registry{},
	}
	testCases := []struct {
		name      string
		builder   *Selector[TestModel]
		wantQuery *Query
		wantErr   error
	}{
		{
			builder: NewSelector[TestModel](db),
			name:    "basic select",
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: []any{},
			},
		},
		{
			name:    "select from",
			builder: NewSelector[TestModel](db).From("`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: []any{},
			},
			wantErr: nil,
		},
		{
			name:    "where",
			builder: NewSelector[TestModel](db).Where(C("Name").Eq("hha")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `name` = ?;",
				Args: []any{"hha"},
			},
			wantErr: nil,
		}, {
			name:    "where not",
			builder: NewSelector[TestModel](db).Where(Not(C("Age").Eq(111))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE NOT (`age` = ?);",
				Args: []any{111},
			},
			wantErr: nil,
		}, {
			name:    "where and",
			builder: NewSelector[TestModel](db).Where((C("Age").Eq(111)).And(C("Name").Eq("hha"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`name` = ?);",
				Args: []any{111, "hha"},
			},
			wantErr: nil,
		}, {
			name:    "where and & not and",
			builder: NewSelector[TestModel](db).Where((C("Age").Eq(111)).And(Not(C("Name").Eq("hha")))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (NOT (`name` = ?));",
				Args: []any{111, "hha"},
			},
			wantErr: nil,
		}, {
			name:    "fromWhere",
			builder: NewSelector[TestModel](db).From("`table_test`").Where((C("Age").Eq(111)).And(Not(C("Name").Eq("hha")))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `table_test` WHERE (`age` = ?) AND (NOT (`name` = ?));",
				Args: []any{111, "hha"},
			},
			wantErr: nil,
		}, {
			name:    "where lt",
			builder: NewSelector[TestModel](db).Where(C("Age").GT(100)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` > ?;",
				Args: []any{100},
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.builder.Build(context.Background())
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, res, tc.wantQuery)
		})
	}
}
