package go_orm

type Expression interface {
	expr()
}

type RawExpression struct {
	expression string
	args       []any
}

func Raw(expr string, args ...any) RawExpression {
	return RawExpression{
		expression: expr,
		args:       args,
	}
}
func (r RawExpression) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
func (r RawExpression) expr()       {}
func (r RawExpression) selectable() {}
