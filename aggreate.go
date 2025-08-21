package go_orm

type Aggregate struct {
	fn    string
	col   Column
	alias string
}

func (a Aggregate) selectable() {}
func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		col:   a.col,
		alias: alias,
	}
}
func Max(column string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		col: C(column),
	}
}

func Min(column string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		col: C(column),
	}
}
func Count(column string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		col: C(column),
	}
}
func CountAll() Aggregate {
	return Aggregate{
		fn:  "COUNT",
		col: C("*"),
	}
}

func Sum(column string) Aggregate {
	return Aggregate{
		fn:  "SUM",
		col: C(column),
	}
}

func Avg(column string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		col: C(column),
	}
}
