package go_orm

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpdater_Build(t *testing.T) {
	type TestModel struct {
		Name string
		Age  int
	}
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB, WithDialect(&mysqlDialect{}))

	testCases := []struct {
		name      string
		builder   *Updater[TestModel]
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "update with assigns single",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.Set(Assignment{
					column: C("Name"),
					val:    "修",
				})
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `name` = ?;",
				Args: []any{"修"},
			},
			wantErr: nil,
		},
		{
			name: "update with assigns multiple",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.Set(
					Assignment{column: C("Name"), val: "Shu"},
					Assignment{column: C("Age"), val: 19},
				)
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `name` = ?, `age` = ?;",
				Args: []any{"Shu", 19},
			},
			wantErr: nil,
		},
		{
			name: "update from struct - skip zero",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.FromStruct(&TestModel{Name: "Shu", Age: 0})
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `name` = ?;",
				Args: []any{"Shu"},
			},
			wantErr: nil,
		},
		{
			name: "update from struct multiple fields",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.FromStruct(&TestModel{Name: "Shu", Age: 19})
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `name` = ?, `age` = ?;",
				Args: []any{"Shu", 19},
			},
			wantErr: nil,
		},
		{
			name: "update no assigns",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				return u
			}(),
			wantQuery: nil,
			wantErr:   ErrUpdateNoColumns,
		},
		{
			name: "update with where single",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.Set(Assignment{column: C("Age"), val: 20})
				u.Where(C("Name").Eq("Shu"))
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `age` = ? WHERE `name` = ?;",
				Args: []any{20, "Shu"},
			},
			wantErr: nil,
		},
		{
			name: "update with where raw",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.Set(Assignment{column: C("Age"), val: 20})
				u.Where(Raw("ID = ?", 18).AsPredicate())
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `age` = ? WHERE ID = ?;",
				Args: []any{20, 18},
			},
			wantErr: nil,
		},
		{
			name: "update with where raw and col",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.Set(Assignment{column: C("Age"), val: 20})
				u.Where(Raw("ID = ?", 18).AsPredicate().And(C("Name").Eq("c")))
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `age` = ? WHERE (ID = ?) AND (`name` = ?);",
				Args: []any{20, 18, "c"},
			},
			wantErr: nil,
		},
		{
			name: "update with where not",
			builder: func() *Updater[TestModel] {
				u := NewUpdater[TestModel](db)
				u.Set(Assignment{column: C("Name"), val: "new"})
				u.Where(Not(C("Age").Eq(99)))
				return u
			}(),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `name` = ? WHERE NOT (`age` = ?);",
				Args: []any{"new", 99},
			},
			wantErr: nil,
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
