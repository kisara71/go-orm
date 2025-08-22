package go_orm

type Assignable interface {
	assign()
}

type Assignment struct {
	column Column
	val    any
}

func Assign(column string, val any) Assignment {
	return Assignment{
		column: C(column),
		val:    val,
	}
}
func (Assignment) assign() {}
