package go_orm

type TableReference interface {
	tableReference()
}

type Table struct {
	entity any
}

func (Table) tableReference() {}
func TableOf(entity any) Table {
	return Table{entity: entity}
}
func (t Table) Join(reference TableReference) JoinBuilder {
	return JoinBuilder{
		left:  t,
		typ:   "JOIN",
		right: reference,
	}
}

func (t Table) LeftJoin(reference TableReference) JoinBuilder {
	return JoinBuilder{
		left:  t,
		typ:   "LEFT JOIN",
		right: reference,
	}
}
func (t Table) RightJoin(reference TableReference) JoinBuilder {
	return JoinBuilder{
		left:  t,
		typ:   "RIGHT JOIN",
		right: reference,
	}
}
