package go_orm

type Assignable interface {
	assign()
}

type Assignment struct {
	column string
	val    any
}

func Assign(column string, val any) Assignment {
	return Assignment{
		column: column,
		val:    val,
	}
}
func (Assignment) assign() {}
