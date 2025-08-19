package go_orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSelector(t *testing.T) {
	type TestModel struct {
		Name string
		Age  int
	}
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB)
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

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB)

	type TestModel struct {
		ID      int64  `orm:" column=id_t"`
		Name    string `orm:"column=name_t"`
		Address sql.NullString
	}
	testCases := []struct {
		name     string
		expect   func(mk sqlmock.Sqlmock)
		selector *Selector[TestModel]
		wantRes  *TestModel
		wantErr  error
	}{
		{
			name: "get success",
			expect: func(mk sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id_t",
					"name_t",
					"address",
				})
				rows.AddRow(1, "wang", "lll")
				mk.ExpectQuery("SELECT \\* FROM `test_model` WHERE .*;").WillReturnRows(rows)
			},
			wantErr: nil,
			wantRes: &TestModel{
				ID:   1,
				Name: "wang",
				Address: sql.NullString{
					String: "lll",
					Valid:  true,
				},
			},
			selector: NewSelector[TestModel](db).Where(C("ID").Eq(1)),
		},
		{
			name: "no record",
			expect: func(mk sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id_t", "name_t", "address"})
				mk.ExpectQuery("SELECT \\* FROM `test_model` WHERE .*;").WillReturnRows(rows)
			},
			wantErr:  ErrNoRecord,
			wantRes:  nil,
			selector: NewSelector[TestModel](db).Where(C("ID").Eq(2)),
		},
		{
			name: "unknown column",
			expect: func(mk sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id_t", "name_t", "unknown_col"})
				rows.AddRow(1, "wang", "something")
				mk.ExpectQuery("SELECT \\* FROM `test_model` WHERE .*;").WillReturnRows(rows)
			},
			wantErr:  ErrUnknownColumn,
			wantRes:  nil,
			selector: NewSelector[TestModel](db).Where(C("ID").Eq(1)),
		},
		{
			name: "scan error",
			expect: func(mk sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id_t", "name_t", "address"})
				rows.AddRow("not_int", "wang", "lll")
				mk.ExpectQuery("SELECT \\* FROM `test_model` WHERE .*;").WillReturnRows(rows)
			},
			wantErr:  ErrScanFailed,
			wantRes:  nil,
			selector: NewSelector[TestModel](db).Where(C("ID").Eq(1)),
		},
		{
			name: "build error",
			expect: func(mk sqlmock.Sqlmock) {
			},
			wantErr: ErrUnknownField,
			wantRes: nil,
			selector: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db).Where(C("sfdsf").Eq(1))
				return s
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.expect(mock)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			res, err := tc.selector.Get(ctx)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, res, tc.wantRes)
		})
	}
}
