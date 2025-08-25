package go_orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kisara71/go-orm/errs"
	"github.com/kisara71/go-orm/middleware"
	"github.com/kisara71/go-orm/middleware/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelector(t *testing.T) {
	type TestModel struct {
		Name string
		Age  int
	}
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB, WithDialect(&mysqlDialect{}))
	db.Use(log.NewDefault().Build())
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
			name: "where raw",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db).Where(Raw("ID = ?", 18).AsPredicate())
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE ID = ?;",
				Args: []any{18},
			},
		}, {
			name: "where raw and name",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db).Where(Raw("ID = ?", 18).AsPredicate().And(C("Name").Eq("c")))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (ID = ?) AND (`name` = ?);",
				Args: []any{18, "c"},
			},
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
		{
			name: "max",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Max("Age"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT MAX(`age`) FROM `test_model`;",
				Args: []any{},
			},
			wantErr: nil,
		},
		{
			name: "max min",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Max("Age"), Min("Age"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT MAX(`age`), MIN(`age`) FROM `test_model`;",
				Args: []any{},
			},
			wantErr: nil,
		},
		{
			name: "max, name, min",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Max("Age"), C("Name"), Min("Age"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT MAX(`age`), `name`, MIN(`age`) FROM `test_model`;",
				Args: []any{},
			},
			wantErr: nil,
		}, {
			name: "count all",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(CountAll())
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(*) FROM `test_model`;",
				Args: []any{},
			},
			wantErr: nil,
		},
		{
			name: "sum age",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Sum("Age"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT SUM(`age`) FROM `test_model`;",
				Args: []any{},
			},
			wantErr: nil,
		}, {
			name: "raw expr",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Raw("DISTINCT `age`"), C("Name"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT DISTINCT `age`, `name` FROM `test_model`;",
				Args: []any{},
			},
		}, {
			name: "as",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(C("Age").As("age_as"), Max("Name").As("name_as"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT `age` AS `age_as`, MAX(`name`) AS `name_as` FROM `test_model`;",
				Args: []any{},
			},
		}, {
			name: "group by",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Sum("Age")).GroupBy(C("Age"), Raw("YEAR(`age`)"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT SUM(`age`) FROM `test_model` GROUP BY `age`, YEAR(`age`);",
				Args: []any{},
			},
			wantErr: nil,
		},
		{
			name: "having with aggregate GT",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Sum("Age")).
					GroupBy(C("Age")).
					Having(Sum("Age").GT(100))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT SUM(`age`) FROM `test_model` GROUP BY `age` HAVING SUM(`age`) > ?;",
				Args: []any{100},
			},
			wantErr: nil,
		},
		{
			name: "having with column condition",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Sum("Age")).
					GroupBy(C("Age")).
					Having(C("Age").GT(30))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT SUM(`age`) FROM `test_model` GROUP BY `age` HAVING `age` > ?;",
				Args: []any{30},
			},
			wantErr: nil,
		},
		{
			name: "having with aggregate AND column",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Sum("Age")).
					GroupBy(C("Age")).
					Having(Sum("Age").GT(100).And(C("Age").LT(50)))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT SUM(`age`) FROM `test_model` GROUP BY `age` HAVING (SUM(`age`) > ?) AND (`age` < ?);",
				Args: []any{100, 50},
			},
			wantErr: nil,
		},
		{
			name: "order",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Raw("DISTINCT `age`"), C("Name")).OrderBy(ASC("Age"), DESC("Age"))
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT DISTINCT `age`, `name` FROM `test_model` ORDER BY `age` ASC, `age` DESC;",
				Args: []any{},
			},
		},
		{
			name: "limit Offset",
			builder: func() *Selector[TestModel] {
				s := NewSelector[TestModel](db)
				s.Select(Raw("DISTINCT `age`"), C("Name")).OrderBy(ASC("Age"), DESC("Age")).
					Limit(1).Offset(10)
				return s
			}(),
			wantQuery: &Query{
				SQL:  "SELECT DISTINCT `age`, `name` FROM `test_model` ORDER BY `age` ASC, `age` DESC LIMIT ? OFFSET ?;",
				Args: []any{int64(1), int64(10)},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &middleware.Context{Ctx: context.Background()}
			err := tc.builder.Build(ctx)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, &Query{
				SQL:  ctx.Statement,
				Args: ctx.Args,
			})
		})
	}
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB, WithDialect(MySQLDialect))
	db.Use(log.NewDefault().Build())
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
			wantErr:  errs.ErrNoRecord,
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
			wantErr:  errs.ErrUnknownColumn,
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
			wantErr:  errs.ErrScanFailed,
			wantRes:  nil,
			selector: NewSelector[TestModel](db).Where(C("ID").Eq(1)),
		},
		{
			name: "build error",
			expect: func(mk sqlmock.Sqlmock) {
			},
			wantErr: errs.ErrUnknownField,
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
			ctx := &middleware.Context{Ctx: context.Background()}
			res, err := tc.selector.Get(ctx)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
func TestSelector_GetMulti(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB, WithDialect(MySQLDialect))
	db.Use(log.NewDefault().Build())
	type TestModel struct {
		ID      int64  `orm:"column=id_t"`
		Name    string `orm:"column=name_t"`
		Address sql.NullString
	}

	testCases := []struct {
		name     string
		expect   func(mk sqlmock.Sqlmock)
		selector *Selector[TestModel]
		wantRes  []*TestModel
		wantErr  error
	}{
		{
			name: "get multi success",
			expect: func(mk sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id_t", "name_t", "address"}).
					AddRow(1, "wang", "addr1").
					AddRow(2, "li", "addr2")
				mk.ExpectQuery("SELECT \\* FROM `test_model` WHERE .*;").WillReturnRows(rows)
			},
			wantErr: nil,
			wantRes: []*TestModel{
				{ID: 1, Name: "wang", Address: sql.NullString{String: "addr1", Valid: true}},
				{ID: 2, Name: "li", Address: sql.NullString{String: "addr2", Valid: true}},
			},
			selector: NewSelector[TestModel](db).Where(C("ID").GT(0)),
		},
		{
			name: "no record",
			expect: func(mk sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id_t", "name_t", "address"})
				mk.ExpectQuery("SELECT \\* FROM `test_model` WHERE .*;").WillReturnRows(rows)
			},
			wantErr:  nil, // GetMulti 返回空 slice 不返回 ErrNoRecord
			wantRes:  []*TestModel{},
			selector: NewSelector[TestModel](db).Where(C("ID").Eq(100)),
		},
		{
			name: "scan error",
			expect: func(mk sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id_t", "name_t", "address"}).
					AddRow("not_int", "wang", "addr")
				mk.ExpectQuery("SELECT \\* FROM `test_model` WHERE .*;").WillReturnRows(rows)
			},
			wantErr:  errs.ErrScanFailed,
			wantRes:  nil,
			selector: NewSelector[TestModel](db).Where(C("ID").Eq(1)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.expect(mock)
			ctx := &middleware.Context{Ctx: context.Background()}
			res, err := tc.selector.GetMulti(ctx)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
