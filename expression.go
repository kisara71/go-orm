package go_orm

import "strings"

type Expression interface {
	expr()
}

func buildExpression(sb *strings.Builder, args *[]any, p Expression, fields map[string]*fieldInfo) error {
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
		if err := buildColumns(t, sb, fields); err != nil {
			return err
		}
		sb.WriteByte(' ')
	case Arg:
		sb.WriteByte('?')
		*args = append(*args, t.val)
	}
	return nil
}

func buildColumns(col Column, sb *strings.Builder, fields map[string]*fieldInfo) error {
	if col.name == "*" {
		sb.WriteByte('*')
		return nil
	}
	if _, ok := fields[col.name]; !ok {
		return ErrUnknownField
	}
	sb.WriteByte('`')
	sb.WriteString(fields[col.name].colName)
	sb.WriteByte('`')
	return nil
}

func buildAggregates(aggregate Aggregate, sb *strings.Builder, fields map[string]*fieldInfo) error {
	sb.WriteString(aggregate.fn)
	sb.WriteByte('(')
	if err := buildColumns(aggregate.col, sb, fields); err != nil {
		return err
	}
	sb.WriteByte(')')
	return nil
}
