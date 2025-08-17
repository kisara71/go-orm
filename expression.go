package go_orm

import "strings"

type Expression interface {
	expr()
}

func buildExpression(sb *strings.Builder, args *[]any, p Expression, fields map[string]fieldInfo) error {
	if p == nil {
		return nil
	}
	switch t := p.(type) {
	case Predicate:
		_, ok := t.left.(Predicate)
		if ok {
			sb.WriteByte('(')
		}
		if err := buildExpression(sb, args, t.left, fields); err != nil {
			return err
		}
		if ok {
			sb.WriteString(") ")
		}
		sb.WriteString(t.op.String())
		sb.WriteByte(' ')
		_, ok = t.right.(Predicate)
		if ok {
			sb.WriteByte('(')
		}
		if err := buildExpression(sb, args, t.right, fields); err != nil {
			return err
		}
		if ok {
			sb.WriteByte(')')
		}
	case Column:
		if _, ok := fields[t.name]; !ok {
			return ErrInvalidField
		}
		sb.WriteByte('`')
		sb.WriteString(fields[t.name].colName)
		sb.WriteString("` ")
	case Arg:
		sb.WriteByte('?')
		*args = append(*args, t.val)
	}
	return nil
}
