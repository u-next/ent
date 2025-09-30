package sqlarray

import "entgo.io/ent/dialect/sql"

// Array wraps the expression with the ARRAY function (GoogleSQL).
func Array(e sql.Querier) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		b.WriteString("ARRAY")
		b.Wrap(func(b *sql.Builder) {
			b.Join(e)
		})
	})
}
