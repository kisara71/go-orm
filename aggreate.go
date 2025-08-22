package go_orm

type Aggregate struct {
	fn    string
	col   Column
	alias string
}

func (Aggregate) expr()       {}
func (Aggregate) selectable() {}

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

func (a Aggregate) LT(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: Arg{val: val},
	}
}
func (a Aggregate) GT(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: Arg{val: val},
	}
}
func (a Aggregate) Eq(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: Arg{val: val},
	}
}
