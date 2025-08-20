package go_orm

type Column struct {
	name string
}

func (c Column) selectable() {}
func (c Column) expr()       {}

func C(name string) Column {
	return Column{
		name: name,
	}
}
