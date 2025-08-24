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

func TestInsertor_Build_SQLite(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := OpenDB(mockDB, WithDialect(SqliteDialect))

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
				SQL: `INSERT INTO "test_model" ("id", "name", "address") VALUES (?, ?, ?);`,
				Args: []any{int64(1), "wang", sql.NullString{
					Valid:  true,
					String: "hhha",
				}},
			},
			builder: NewInsertor[TestModel](db).Values(&TestModel{
				ID:      1,
				Name:    "wang",
				Address: sql.NullString{Valid: true, String: "hhha"},
			}),
		},
		{
			name: "column insert",
			wantQuery: &Query{
				SQL:  `INSERT INTO "test_model" ("id", "address") VALUES (?, ?);`,
				Args: []any{int64(1), sql.NullString{Valid: true, String: "hhha"}},
			},
			builder: NewInsertor[TestModel](db).Columns("ID", "Address").Values(&TestModel{
				ID:      1,
				Address: sql.NullString{Valid: true, String: "hhha"},
			}),
		},
		{
			name: "multi insert",
			wantQuery: &Query{
				SQL: `INSERT INTO "test_model" ("id", "name", "address") VALUES (?, ?, ?), (?, ?, ?);`,
				Args: []any{
					int64(1), "wang", sql.NullString{Valid: true, String: "hhha"},
					int64(2), "li", sql.NullString{Valid: true, String: "ahhh"},
				},
			},
			builder: NewInsertor[TestModel](db).
				Values(&TestModel{ID: 1, Name: "wang", Address: sql.NullString{Valid: true, String: "hhha"}}).
				Values(&TestModel{ID: 2, Name: "li", Address: sql.NullString{Valid: true, String: "ahhh"}}),
		},
		{
			name:    "no values",
			wantErr: ErrInsertNoValues,
			builder: NewInsertor[TestModel](db).Values(),
		},
		{
			name: "on conflict update assign",
			wantQuery: &Query{
				SQL: `INSERT INTO "test_model" ("id", "name", "address") VALUES (?, ?, ?) ` +
					`ON CONFLICT("id") DO UPDATE SET "name" = ?;`,
				Args: []any{
					int64(1), "wang", sql.NullString{Valid: true, String: "hhha"},
					"shi",
				},
			},
			builder: NewInsertor[TestModel](db).Values(&TestModel{
				ID:      1,
				Name:    "wang",
				Address: sql.NullString{Valid: true, String: "hhha"},
			}).OnConflict().Columns(C("ID")).Update(Assign("Name", "shi")),
		},
		{
			name: "on conflict update assign and column",
			wantQuery: &Query{
				SQL: `INSERT INTO "test_model" ("id", "name", "address") VALUES (?, ?, ?) ` +
					`ON CONFLICT("id") DO UPDATE SET "name" = ?, "address" = excluded."address";`,
				Args: []any{
					int64(1), "wang", sql.NullString{Valid: true, String: "hhha"},
					"shi",
				},
			},
			builder: NewInsertor[TestModel](db).Values(&TestModel{
				ID:      1,
				Name:    "wang",
				Address: sql.NullString{Valid: true, String: "hhha"},
			}).OnConflict().Columns(C("ID")).Update(Assign("Name", "shi"), C("Address")),
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
