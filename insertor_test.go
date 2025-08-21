package go_orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInsertor_Build(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB)
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
			}).OnDuplicateKey().Update(Assign("Name", "shi")),
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
			}).OnDuplicateKey().Update(Assign("Name", "shi"), C("Address")),
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
