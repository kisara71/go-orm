package go_orm

type op string

const (
	opEq  op = "="
	opAnd op = "AND"
	opNot op = "NOT"
	opOr  op = "OR"
	opLT  op = "<"
	opGT  op = ">"
)

type Clause int

const (
	ClauseWhere Clause = iota
	ClauseHaving
	ClauseOn
)

func (o op) String() string {
	return string(o)
}

func (p Predicate) expr() {}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

type Arg struct {
	val any
}

func (a Arg) expr() {}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: Arg{val: arg},
	}
}

func (c Column) LT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: Arg{val: val},
	}
}
func (c Column) GT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: Arg{val: val},
	}
}
func Not(predicate Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: predicate,
	}
}

func (p Predicate) And(predicate Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opAnd,
		right: predicate,
	}
}
func (p Predicate) Or(predicate Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opOr,
		right: predicate,
	}
}
