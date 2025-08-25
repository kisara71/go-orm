package go_orm

import (
	"github.com/kisara71/go-orm/errs"
	"github.com/kisara71/go-orm/model"
	"strings"
)

type builder struct {
	m       *model.Model
	sb      strings.Builder
	args    []any
	dialect Dialect
	qte     byte
}

func NewBuilder(m *model.Model, dialect Dialect) *builder {
	return &builder{
		m:       m,
		sb:      strings.Builder{},
		args:    make([]any, 0, 8),
		dialect: dialect,
		qte:     dialect.Quoter(),
	}
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.qte)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.qte)
}
func (b *builder) getSQL() string {
	return b.sb.String()
}
func (b *builder) getArgs() []any {
	return b.args
}
func (b *builder) buildString(val string) {
	b.sb.WriteString(val)
}
func (b *builder) buildByte(val byte) {
	b.sb.WriteByte(val)
}

func (b *builder) buildColumn(col Column) error {
	if col.name == "*" {
		b.sb.WriteByte('*')
		return nil
	}
	if _, ok := b.m.GoMap[col.name]; !ok {
		return errs.ErrUnknownField
	}
	b.quote(b.m.GoMap[col.name].ColName)
	if col.alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(col.alias)
	}
	return nil
}

func (b *builder) buildExpression(exp Expression, clause Clause) error {
	if exp == nil {
		return nil
	}
	switch t := exp.(type) {
	case Predicate:
		_, ok := t.left.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(t.left, clause); err != nil {
			return err
		}
		if ok {
			b.sb.WriteString(") ")
		}
		if t.op != "" {
			b.sb.WriteString(t.op.String())
			b.sb.WriteByte(' ')
		}
		_, ok = t.right.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(t.right, clause); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
	case Column:
		if err := b.buildColumn(t); err != nil {
			return err
		}
		b.sb.WriteByte(' ')
	case Arg:
		b.sb.WriteByte('?')
		b.addArgs(t.val)
	case RawExpression:
		b.sb.WriteString(t.expression)
		b.addArgs(t.args...)
	case Aggregate:
		if clause == ClauseWhere || clause == ClauseOn {
			return errs.ErrUnsupportedType
		}
		if err := b.buildAggregate(t); err != nil {
			return err
		}
		b.sb.WriteByte(' ')
	}

	return nil
}

func (b *builder) buildAggregate(aggregate Aggregate) error {
	b.sb.WriteString(aggregate.fn)
	b.sb.WriteByte('(')
	if err := b.buildColumn(aggregate.col); err != nil {
		return err
	}
	b.sb.WriteByte(')')
	if aggregate.alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(aggregate.alias)
	}
	return nil
}

func (b *builder) addArgs(arg ...any) {
	if len(arg) == 0 {
		return
	}
	b.args = append(b.args, arg...)
}
