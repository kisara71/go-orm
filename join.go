package go_orm

type JoinBuilder struct {
	left  TableReference
	typ   string
	right TableReference
}
type join interface {
	join()
}

func (JoinBuilder) join()           {}
func (JoinBuilder) tableReference() {}
func (j JoinBuilder) On(predicate ...Predicate) Join {
	return Join{
		left:  j.left,
		typ:   j.typ,
		right: j.right,
		on:    predicate,
	}
}
func (j JoinBuilder) Using(columns ...string) Join {
	return Join{
		left:  j.left,
		typ:   j.typ,
		right: j.right,
		using: columns,
	}
}
func (j JoinBuilder) toJoin() Join {
	return Join{
		left:  j.left,
		typ:   j.typ,
		right: j.right,
	}
}

type Join struct {
	left  TableReference
	typ   string
	right TableReference
	on    []Predicate
	using []string
}

func (Join) join()           {}
func (Join) tableReference() {}
func (j Join) Join(reference TableReference) JoinBuilder {
	return JoinBuilder{
		left:  j,
		typ:   "JOIN",
		right: reference,
	}
}

func (j Join) LeftJoin(reference TableReference) JoinBuilder {
	return JoinBuilder{
		left:  j,
		typ:   "LEFT JOIN",
		right: reference,
	}
}
func (j Join) RightJoin(reference TableReference) JoinBuilder {
	return JoinBuilder{
		left:  j,
		typ:   "RIGHT JOIN",
		right: reference,
	}
}
