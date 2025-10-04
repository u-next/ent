package sqlarray

import "entgo.io/ent/dialect/sql"

// Array wraps the expression with the ARRAY function (GoogleSQL).
func Array(e string) func(*sql.Builder) {
	return func(b *sql.Builder) {
		b.WriteString("ARRAY(")
		b.WriteString(e)
		b.WriteString(")")
	}
}

// UnnestBuilder is a builder for the UNNEST expression.
type UnnestBuilder struct {
	sql.TableView
	expr   sql.Querier
	alias  string
	offset string
}

// Unnest wraps the array expression with the UNNEST operator (GoogleSQL).
func Unnest(array sql.Querier) *UnnestBuilder {
	return &UnnestBuilder{
		expr: array,
	}
}

func (u *UnnestBuilder) As(alias string) *UnnestBuilder {
	u.alias = alias
	return u
}

func (u *UnnestBuilder) WithOffset() *UnnestBuilder {
	u.offset = "offset"
	return u
}

func (u *UnnestBuilder) WithOffsetAs(offset string) *UnnestBuilder {
	u.offset = offset
	return u
}

func (u *UnnestBuilder) Query() string {
	b := &sql.Builder{}
	b.WriteString("UNNEST ")
	b.Wrap(func(b *sql.Builder) {
		b.Join(u.expr)
	})
	if u.alias != "" {
		b.WriteString(" AS ")
		b.WriteString(u.alias)
	}
	if u.offset != "" {
		b.WriteString(" WITH OFFSET")
		if u.offset != "offset" {
			b.WriteString(" AS ")
			b.WriteString(u.offset)
		}
	}
	return b.String()
}
