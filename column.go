package go_orm

type Column struct {
	name  string
	alias string
}

func (c Column) selectable() {}
func (c Column) expr()       {}

func C(name string) Column {
	return Column{
		name: name,
	}
}
func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}
