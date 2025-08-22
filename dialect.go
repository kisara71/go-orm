package go_orm

type Dialect interface {
	Quoter() byte
	BuildUpsert(builder *builder, opk *OnConflict) error
}

var (
	StandardSQL    = &standardSQL{}
	MySQLDialect   = &mysqlDialect{standardSQL: StandardSQL}
	SqliteDialect  = &sqliteDialect{standardSQL: StandardSQL}
	PostGreDialect = &postgreDialect{standardSQL: StandardSQL}
)

type standardSQL struct {
}

func (s *standardSQL) Quoter() byte {
	return '"'
}

func (s *standardSQL) BuildUpsert(builder *builder, opk *OnConflict) error {
	return ErrUnsupported
}

type mysqlDialect struct {
	*standardSQL
}

func (m *mysqlDialect) Quoter() byte {
	return '`'
}
func (m *mysqlDialect) BuildUpsert(builder *builder, opk *OnConflict) error {
	builder.buildString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range opk.assigns {
		if idx > 0 {
			builder.buildString(", ")
		}
		switch as := assign.(type) {
		case Assignment:
			if err := builder.buildColumn(as.column); err != nil {
				return err
			}
			builder.buildString(" = ?")
			builder.addArgs(as.val)
		case Column:
			err := builder.buildColumn(as)
			if err != nil {
				return ErrUnknownField
			}
			builder.buildString(" = VALUES(")
			_ = builder.buildColumn(as)
			builder.buildByte(')')
		}
	}
	return nil
}

type sqliteDialect struct {
	*standardSQL
}

func (s *sqliteDialect) BuildUpsert(builder *builder, opk *OnConflict) error {
	builder.buildString(" ON CONFLICT(")
	for i, col := range opk.conflictColumns {
		if i > 0 {
			builder.buildString(", ")
		}
		if err := builder.buildColumn(col); err != nil {
			return err
		}
	}
	builder.buildString(") DO UPDATE SET ")

	for i, assign := range opk.assigns {
		if i > 0 {
			builder.buildString(", ")
		}
		switch as := assign.(type) {
		case Assignment:
			if err := builder.buildColumn(as.column); err != nil {
				return err
			}
			builder.buildString(" = ?")
			builder.addArgs(as.val)
		case Column:
			err := builder.buildColumn(as)
			if err != nil {
				return ErrUnknownField
			}
			builder.buildString(" = excluded.")
			_ = builder.buildColumn(as)
		}
	}
	return nil
}

type postgreDialect struct {
	*standardSQL
}

func (p *postgreDialect) BuildUpsert(builder *builder, opk *OnConflict) error {
	builder.buildString(" ON CONFLICT(")
	for i, col := range opk.conflictColumns {
		if i > 0 {
			builder.buildString(", ")
		}
		if err := builder.buildColumn(col); err != nil {
			return err
		}
	}
	builder.buildString(") DO UPDATE SET ")

	for i, assign := range opk.assigns {
		if i > 0 {
			builder.buildString(", ")
		}
		switch as := assign.(type) {
		case Assignment:
			if err := builder.buildColumn(as.column); err != nil {
				return err
			}
			builder.buildString(" = ?")
			builder.addArgs(as.val)
		case Column:
			err := builder.buildColumn(as)
			if err != nil {
				return ErrUnknownField
			}
			builder.buildString(" = excluded.")
			_ = builder.buildColumn(as)
		}
	}
	return nil
}
