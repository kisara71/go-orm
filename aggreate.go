package go_orm

type Aggregate struct {
	fn  string
	col Column
}

func (a Aggregate) selectable() {}

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
