package go_orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kisara71/go-orm/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInsertor_Mysql_Build(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB, WithDialect(&mysqlDialect{}))
	type TestModel struct {
		ID      int64
		Name    string
		Address sql.NullString
	}
	testCases := []struct {
		name      string
		wantQuery *Query
		wantErr   error
		builder   Builder
	}{
		{
			name: "default insert",
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`, `name`, `address`) VALUES (?, ?, ?);",
				Args: []any{int64(1), "wang", sql.NullString{
					Valid:  true,
					String: "hhha",
				}},
			},
			wantErr: nil,
			builder: NewInsertor[TestModel](db).Values(&TestModel{
				ID:   1,
				Name: "wang",
				Address: sql.NullString{
					Valid:  true,
					String: "hhha",
				},
			}),
		},
		{
			name: "column insert",
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`, `address`) VALUES (?, ?);",
				Args: []any{int64(1), sql.NullString{
					Valid:  true,
					String: "hhha",
				}},
			},
			wantErr: nil,
			builder: NewInsertor[TestModel](db).Columns("ID", "Address").Values(&TestModel{
				ID: 1,
				Address: sql.NullString{
					Valid:  true,
					String: "hhha",
				},
			}),
		},
		{
			name: "multi insert",
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`, `name`, `address`) VALUES (?, ?, ?), (?, ?, ?);",
				Args: []any{int64(1), "wang", sql.NullString{
					Valid:  true,
					String: "hhha",
				}, int64(2), "li", sql.NullString{
					Valid:  true,
					String: "ahhh",
				}},
			},
			wantErr: nil,
			builder: NewInsertor[TestModel](db).Values(&TestModel{
				ID:   1,
				Name: "wang",
				Address: sql.NullString{
					Valid:  true,
					String: "hhha",
				},
			}, &TestModel{
				ID:   2,
				Name: "li",
				Address: sql.NullString{
					Valid:  true,
					String: "ahhh",
				},
			},
			),
		},
		{
			name:      "no values",
			wantQuery: nil,
			wantErr:   ErrInsertNoValues,
			builder:   NewInsertor[TestModel](db).Values(),
		},
		{
			name: "on duplicate update assign",
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`, `name`, `address`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE" +
					" `name` = ?;",
				Args: []any{int64(1), "wang", sql.NullString{
					Valid:  true,
					String: "hhha",
				}, "shi"},
			},
			wantErr: nil,
			builder: NewInsertor[TestModel](db).Values(&TestModel{
				ID:   1,
				Name: "wang",
				Address: sql.NullString{
					Valid:  true,
					String: "hhha",
				},
			}).OnConflict().Update(Assign("Name", "shi")),
		},
		{
			name: "on duplicate update assign and column",
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`, `name`, `address`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE" +
					" `name` = ?, `address` = VALUES(`address`);",
				Args: []any{int64(1), "wang", sql.NullString{
					Valid:  true,
					String: "hhha",
				}, "shi"},
			},
			wantErr: nil,
			builder: NewInsertor[TestModel](db).Values(&TestModel{
				ID:   1,
				Name: "wang",
				Address: sql.NullString{
					Valid:  true,
					String: "hhha",
				},
			}).OnConflict().Update(Assign("Name", "shi"), C("Address")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &middleware.Context{Ctx: context.Background()}
			err := tc.builder.Build(ctx)
			assert.Equal(t, tc.wantErr, err)
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
func TestInsertor_Exec_MySQL(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := OpenDB(mockDB, WithDialect(&mysqlDialect{}))

	type TestModel struct {
		ID      int64
		Name    string
		Address sql.NullString
	}

	testCases := []struct {
		name         string
		values       []*TestModel
		onDuplicate  []Assignable
		mockExpect   func()
		wantLastID   int64
		wantAffected int64
		wantErr      error
	}{
		{
			name: "single insert",
			values: []*TestModel{
				{ID: 1, Name: "wang"},
			},
			mockExpect: func() {
				mock.ExpectExec("INSERT INTO `test_model`").
					WithArgs(int64(1), "wang", sql.NullString{}).
					WillReturnResult(sqlmock.NewResult(101, 1))
			},
			wantLastID:   101,
			wantAffected: 1,
			wantErr:      nil,
		},
		{
			name: "multi insert",
			values: []*TestModel{
				{ID: 1, Name: "wang"},
				{ID: 2, Name: "li"},
			},
			mockExpect: func() {
				mock.ExpectExec("INSERT INTO `test_model`").
					WithArgs(int64(1), "wang", sql.NullString{},
						int64(2), "li", sql.NullString{}).
					WillReturnResult(sqlmock.NewResult(102, 2))
			},
			wantLastID:   102,
			wantAffected: 2,
			wantErr:      nil,
		},
		{
			name: "on duplicate key update",
			values: []*TestModel{
				{ID: 1, Name: "wang"},
			},
			onDuplicate: []Assignable{Assign("Name", "shi")},
			mockExpect: func() {
				mock.ExpectExec("INSERT INTO `test_model`").
					WithArgs(int64(1), "wang", sql.NullString{}, "shi").
					WillReturnResult(sqlmock.NewResult(103, 1))
			},
			wantLastID:   103,
			wantAffected: 1,
			wantErr:      nil,
		},
		{
			name:         "no values",
			values:       nil,
			mockExpect:   nil,
			wantLastID:   0,
			wantAffected: 0,
			wantErr:      ErrInsertNoValues,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			insertor := NewInsertor[TestModel](db)
			if len(tc.values) > 0 {
				insertor = insertor.Values(tc.values...)
			}
			if len(tc.onDuplicate) > 0 {
				insertor = insertor.OnConflict().Update(tc.onDuplicate...)
			}

			if tc.mockExpect != nil {
				tc.mockExpect()
			}

			ctx := &middleware.Context{Ctx: context.Background()}
			res := insertor.Exec(ctx)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, res.Err())
				return
			}
			assert.NoError(t, res.Err())
			lastID, _ := res.LastInsertID()
			assert.Equal(t, tc.wantLastID, lastID)
			rowsAffected, _ := res.RowsAffected()
			assert.Equal(t, tc.wantAffected, rowsAffected)

			err := mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
