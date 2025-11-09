package sqlpgq

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlhint"
)

// GraphQuery is a builder for property graph queries.
type GraphQuery struct {
	sql.Builder
	graph string        // property graph name
	stmts []sql.Querier // linear query statements
}

// Graph returns a new GraphQuery for the GRAPH statement.
func Graph(name string) *GraphQuery {
	return &GraphQuery{
		graph: name,
	}
}

// GraphExpr returns a new graph expression.
func GraphExpr() *GraphQuery {
	return &GraphQuery{}
}

// Append adds a linear query statement to the multi-linear query.
func (g *GraphQuery) Append(stmt Statement) *GraphQuery {
	g.stmts = append(g.stmts, stmt)
	return g
}

// Match adds a MATCH statement to the linear query.
func (g *GraphQuery) Match(patterns ...Pattern) *GraphQuery {
	return g.Append(Match(patterns...))
}

// MatchWithHint adds a MATCH statement with a hint to the linear query.
func (g *GraphQuery) MatchWithHint(hint sqlhint.JoinHint, patterns ...Pattern) *GraphQuery {
	return g.Append(Match(patterns...).Hint(hint))
}

// OptionalMatch adds an OPTIONAL MATCH statement to the linear query.
func (g *GraphQuery) OptionalMatch(patterns ...Pattern) *GraphQuery {
	return g.Append(OptionalMatch(patterns...))
}

// OptionalMatchWithHint adds an OPTIONAL MATCH statement with a hint to the linear query.
func (g *GraphQuery) OptionalMatchWithHint(hint sqlhint.JoinHint, patterns ...Pattern) *GraphQuery {
	return g.Append(OptionalMatch(patterns...).Hint(hint))
}

// Filter adds a FILTER statement to the linear query.
func (g *GraphQuery) Filter(p *sql.Predicate) *GraphQuery {
	return g.Append(Filter(p))
}

// For adds a FOR statement to the linear query.
func (g *GraphQuery) For(f *ForBuilder) *GraphQuery {
	return g.Append(f)
}

// Let adds a LET statement to the linear query.
func (g *GraphQuery) Let(as ...*assignment) *GraphQuery {
	l := Let().Append(as...)
	return g.Append(l)
}

// Return adds a RETURN statement to the linear query.
func (g *GraphQuery) Return(items ...string) *GraphQuery {
	if len(g.stmts) == 0 {
		return g.Append(Return(items...))
	}
	last := g.stmts[len(g.stmts)-1]
	switch last := last.(type) {
	case *ReturnBuilder:
		last.AppendItem(items...)
	default:
		g.Append(Return(items...))
	}
	return g
}

// ReturnExprAs adds a return expression with an alias to the last RETURN statement.
func (g *GraphQuery) ReturnExprAs(expr sql.Querier, as string) *GraphQuery {
	if len(g.stmts) == 0 {
		return g.Append(Return().AppendItemExprAs(expr, as))
	}
	last := g.stmts[len(g.stmts)-1]
	switch last := last.(type) {
	case *ReturnBuilder:
		last.AppendItemExprAs(expr, as)
	default:
		g.Append(Return().AppendItemExprAs(expr, as))
	}
	return g
}

// ReturnAs adds a return item with an alias to the last RETURN statement.
func (g *GraphQuery) ReturnAs(it, as string) *GraphQuery {
	if len(g.stmts) == 0 {
		return g.Append(Return().AppendItemAs(it, as))
	}
	last := g.stmts[len(g.stmts)-1]
	switch last := last.(type) {
	case *ReturnBuilder:
		last.AppendItemAs(it, as)
	default:
		g.Append(Return().AppendItemAs(it, as))
	}
	return g
}

// ReturnDistinct adds a RETURN DISTINCT statement to the linear query.
func (g *GraphQuery) ReturnDistinct(items ...string) *GraphQuery {
	return g.Append(Return().ReturnDistinct(items...))
}

// GroupBy adds a GROUP BY clause to the last RETURN or WITH statement.
func (g *GraphQuery) GroupBy(items ...string) *GraphQuery {
	return g.Append(GroupBy(items...))
}

// OrderBy adds an ORDER BY clause to the last RETURN statement or the linear query.
func (g *GraphQuery) OrderBy(columns ...string) *GraphQuery {
	return g.Append(OrderBy(columns...))
}

// OrderByExpr adds the `ORDER BY` clause to the last RETURN statement or the linear query.
func (g *GraphQuery) OrderByExpr(exprs ...sql.Querier) *GraphQuery {
	return g.Append(OrderByExpr(exprs...))
}

// Limit adds a LIMIT statement to the linear query.
func (g *GraphQuery) Limit(count int) *GraphQuery {
	return g.Append(Limit(count))
}

// Offset adds an OFFSET statement to the linear query.
func (g *GraphQuery) Offset(count int) *GraphQuery {
	return g.Append(Offset(count))
}

// Skip is an alias for Offset that uses the SKIP keyword.
func (g *GraphQuery) Skip(count int) *GraphQuery {
	return g.Append(Skip(count))
}

// With adds a WITH statement to the linear query.
func (g *GraphQuery) With(items ...string) *GraphQuery {
	return g.Append(With(items...))
}

// WithAll adds a WITH ALL statement to the linear query.
func (g *GraphQuery) WithAll(items ...string) *GraphQuery {
	return g.Append(WithAll(items...))
}

// WithDistinct adds a WITH DISTINCT statement to the linear query.
func (g *GraphQuery) WithDistinct(items ...string) *GraphQuery {
	return g.Append(WithDistinct(items...))
}

// Next chains multiple linear queries with NEXT statements.
func (g *GraphQuery) Next() *GraphQuery {
	return g.Append(Next)
}

// UnionAll adds a UNION ALL operation to the composite query.
func (g *GraphQuery) UnionAll() *GraphQuery {
	return g.Append(UnionAll)
}

// UnionDistinct adds a UNION DISTINCT operation to the composite query.
func (g *GraphQuery) UnionDistinct() *GraphQuery {
	return g.Append(UnionDistinct)
}

// IntersectAll adds an INTERSECT ALL operation to the composite query.
func (g *GraphQuery) IntersectAll() *GraphQuery {
	return g.Append(IntersectAll)
}

// IntersectDistinct adds an INTERSECT DISTINCT operation to the composite query.
func (g *GraphQuery) IntersectDistinct() *GraphQuery {
	return g.Append(IntersectDistinct)
}

// ExceptAll adds an EXCEPT ALL operation to the composite query.
func (g *GraphQuery) ExceptAll() *GraphQuery {
	return g.Append(ExceptAll)
}

// ExceptDistinct adds an EXCEPT DISTINCT operation to the composite query.
func (g *GraphQuery) ExceptDistinct() *GraphQuery {
	return g.Append(ExceptDistinct)
}

// Query returns the property graph query representation.
func (g *GraphQuery) Query() (string, []any) {
	if g.graph != "" {
		g.WriteString("GRAPH ")
		g.Ident(g.graph)
		g.NewLine()
	}
	if len(g.stmts) > 0 {
		g.JoinNewLine(g.stmts...)
	}
	return g.String(), g.GetArgs()
}

// Statement is a linear query statement.
type Statement interface {
	sql.Querier
	stmt()
}

// Matcher is a builder for MATCH statements.
type Matcher struct {
	sql.Builder
	optional bool
	hint     sqlhint.JoinHint
	patterns []sql.Querier
}

// Match creates a new MATCH statement builder.
func Match(patterns ...Pattern) *Matcher {
	return (&Matcher{}).AppendPatterns(patterns...)
}

// OptionalMatch creates a new OPTIONAL MATCH statement builder.
func OptionalMatch(patterns ...Pattern) *Matcher {
	return (&Matcher{optional: true}).AppendPatterns(patterns...)
}

// Hint sets the match hint.
func (m *Matcher) Hint(hint sqlhint.JoinHint) *Matcher {
	m.hint = hint
	return m
}

// Append adds more patterns to the MATCH statement.
func (m *Matcher) AppendPatterns(patterns ...Pattern) *Matcher {
	for _, p := range patterns {
		m.patterns = append(m.patterns, p)
	}
	return m
}

// Query returns the MATCH statement representation.
func (m *Matcher) Query() (string, []any) {
	if m.optional {
		m.WriteString("OPTIONAL ")
	}
	m.WriteString("MATCH")
	if m.hint != nil {
		m.Pad()
		m.hint.Write(&m.Builder)
	}
	if len(m.patterns) > 0 {
		m.Pad()
		m.JoinComma(m.patterns...)
	}
	return m.String(), m.GetArgs()
}

func (m *Matcher) stmt() {}

// FilterBuilder is a builder for FILTER statements.
type FilterBuilder struct {
	sql.Builder
	where     bool
	predicate *sql.Predicate
	collected [][]*sql.Predicate
}

// Filter creates a new FILTER statement builder.
func Filter(pred *sql.Predicate) *FilterBuilder {
	return &FilterBuilder{predicate: pred}
}

func (f *FilterBuilder) Where() *FilterBuilder {
	f.where = true
	return f
}

// Predicate sets the boolean predicate expression to filter by.
func (f *FilterBuilder) Predicate(pred *sql.Predicate) *FilterBuilder {
	if len(f.collected) > 0 {
		f.collected[len(f.collected)-1] = append(f.collected[len(f.collected)-1], pred)
		return f
	}
	if f.predicate == nil {
		f.predicate = pred
	} else {
		f.predicate = sql.And(f.predicate, pred)
	}
	return f
}

// Query returns the FILTER statement representation.
func (f *FilterBuilder) Query() (string, []any) {
	f.WriteString("FILTER")
	if f.where {
		f.WriteString(" WHERE")
	}
	if f.predicate != nil {
		f.Pad()
		f.Join(f.predicate)
	}
	return f.String(), f.GetArgs()
}

// CollectPredicates indicates the appended predicates should be collected
// and not set as the main predicate.
func (f *FilterBuilder) CollectPredicates() *FilterBuilder {
	f.collected = append(f.collected, []*sql.Predicate{})
	return f
}

// CollectedPredicates returns the collected predicates.
func (f *FilterBuilder) CollectedPredicates() []*sql.Predicate {
	if len(f.collected) == 0 {
		return nil
	}
	return f.collected[len(f.collected)-1]
}

// UncollectedPredicates stop collecting predicates.
func (f *FilterBuilder) UncollectedPredicates() *FilterBuilder {
	if len(f.collected) > 0 {
		f.collected = f.collected[:len(f.collected)-1]
	}
	return f
}

func (f *FilterBuilder) stmt() {}

// item represents a column or an expression item.
type item struct {
	x  sql.Querier
	c  string
	as string
}

// ReturnBuilder is a builder for RETURN statements.
type ReturnBuilder struct {
	sql.Builder
	distinct bool
	items    []item
	groupBy  []sql.Querier
	orderBy  []sql.Querier
	limit    int64
	offset   int64
}

// Return creates a new RETURN statement builder.
func Return(items ...string) *ReturnBuilder {
	return (&ReturnBuilder{}).Return(items...)
}

// Return changes the return items to the given columns.
// If no items are given, it defaults to returning all columns (*).
func (r *ReturnBuilder) Return(items ...string) *ReturnBuilder {
	r.items = make([]item, len(items))
	for i := range items {
		r.items[i] = item{c: items[i]}
	}
	return r
}

// ReturnDistinct returns distinct items.
func (r *ReturnBuilder) ReturnDistinct(items ...string) *ReturnBuilder {
	return r.Return(items...).Distinct()
}

// AppendItem adds more return items to the RETURN statement.
func (r *ReturnBuilder) AppendItem(items ...string) *ReturnBuilder {
	for i := range items {
		r.items = append(r.items, item{c: items[i]})
	}
	return r
}

// AppendItemAs adds a return item to the RETURN statement with the given alias.
func (r *ReturnBuilder) AppendItemAs(it, as string) *ReturnBuilder {
	r.items = append(r.items, item{c: it, as: as})
	return r
}

// AppendItemExprAs adds a return expression to the RETURN statement with the given alias.
func (r *ReturnBuilder) AppendItemExprAs(expr sql.Querier, as string) *ReturnBuilder {
	r.items = append(r.items, item{x: expr, as: as})
	return r
}

// GroupBy adds a GROUP BY clause.
func (r *ReturnBuilder) GroupBy(exprs ...sql.Querier) *ReturnBuilder {
	r.groupBy = append(r.groupBy, exprs...)
	return r
}

// OrderBy adds an ORDER BY clause.
func (r *ReturnBuilder) OrderBy(expr ...sql.Querier) *ReturnBuilder {
	r.orderBy = append(r.orderBy, expr...)
	return r
}

// Distinct adds the DISTINCT keyword to the `SELECT` statement.
func (r *ReturnBuilder) Distinct() *ReturnBuilder {
	r.distinct = true
	return r
}

// SetDistinct sets the DISTINCT keyword to the `SELECT` statement.
func (r *ReturnBuilder) SetDistinct(v bool) *ReturnBuilder {
	r.distinct = v
	return r
}

// Limit adds a LIMIT clause.
func (r *ReturnBuilder) Limit(count int64) *ReturnBuilder {
	r.limit = count
	return r
}

// Offset adds an OFFSET clause.
func (r *ReturnBuilder) Offset(count int64) *ReturnBuilder {
	r.offset = count
	return r
}

// Query returns the RETURN statement representation.
func (r *ReturnBuilder) Query() (string, []any) {
	r.SetDialect(dialect.Spanner)
	r.WriteString("RETURN")
	if r.distinct {
		r.WriteString(" DISTINCT")
	}
	if len(r.items) == 0 {
		r.WriteString(" *")
	} else if len(r.items) > 0 {
		r.Pad()
		for i, item := range r.items {
			if i > 0 {
				r.Comma()
			}
			if item.x != nil {
				r.Join(item.x)
			} else {
				r.WriteString(item.c)
			}
			if item.as != "" {
				r.WriteString(" AS ")
				r.Ident(item.as)
			}
		}
	}
	if len(r.groupBy) > 0 {
		r.NewLine()
		r.WriteString("GROUP BY ")
		r.JoinComma(r.groupBy...)
	}
	if len(r.orderBy) > 0 {
		r.NewLine()
		r.WriteString("ORDER BY ")
		r.JoinComma(r.orderBy...)
	}
	if r.limit != 0 {
		r.NewLine()
		r.WriteString("LIMIT ")
		r.WriteString(strconv.FormatInt(r.limit, 10))
	}
	if r.offset != 0 {
		r.NewLine()
		r.WriteString("OFFSET ")
		r.WriteString(strconv.FormatInt(r.offset, 10))
	}
	return r.String(), r.GetArgs()
}

func (r *ReturnBuilder) stmt() {}

// assignment represents a variable assignment in a LET statement.
type assignment struct {
	variable Var
	value    sql.Querier
}

// Assign creates a new assignment expression.
func Assign(name string, value sql.Querier) *assignment {
	return &assignment{
		variable: Var(name),
		value:    value,
	}
}

// Var sets the variable name to assign to.
func (a *assignment) Var(name string) *assignment {
	a.variable = Var(name)
	return a
}

// Value sets the expression to assign.
func (a *assignment) Value(value sql.Querier) *assignment {
	a.value = value
	return a
}

// F returns the field access expression for the assigned variable.
func (a *assignment) F(fields ...string) string {
	return a.variable.F(fields...)
}

// LetBuilder is a builder for LET statements.
type LetBuilder struct {
	sql.Builder
	assignments []*assignment
}

// Let creates a new LET statement builder.
func Let() *LetBuilder {
	return &LetBuilder{}
}

// Assign adds a new assignment to the LET statement.
func (let *LetBuilder) Assign(name string, value sql.Querier) *LetBuilder {
	let.assignments = append(let.assignments, Assign(name, value))
	return let
}

// Append adds more assignments to the LET statement.
func (l *LetBuilder) Append(as ...*assignment) *LetBuilder {
	l.assignments = append(l.assignments, as...)
	return l
}

// Query returns the LET statement representation.
func (l *LetBuilder) Query() (string, []any) {
	l.WriteString("LET")
	if len(l.assignments) > 0 {
		l.Pad()
		for i, a := range l.assignments {
			if i > 0 {
				l.Comma()
			}
			l.Join(a.variable)
			l.WriteString(" = ")
			l.Join(a.value)
		}
	}
	return l.String(), l.GetArgs()
}

func (l *LetBuilder) stmt() {}

// GroupByBuilder is a builder for GROUP BY statements.
type GroupByBuilder struct {
	sql.Builder
	items []string
}

// GroupBy creates a new GROUP BY statement builder.
func GroupBy(items ...string) *GroupByBuilder {
	return (&GroupByBuilder{}).Append(items...)
}

// Append adds more expressions to the GROUP BY statement.
func (g *GroupByBuilder) Append(items ...string) *GroupByBuilder {
	g.items = append(g.items, items...)
	return g
}

// Query returns the GROUP BY statement representation.
func (g *GroupByBuilder) Query() (string, []any) {
	g.WriteString("GROUP BY")
	if len(g.items) > 0 {
		g.Pad()
		g.WriteString(strings.Join(g.items, ", "))
	}
	return g.String(), g.GetArgs()
}

func (g *GroupByBuilder) stmt() {}

// OrderByBuilder is a builder for ORDER BY statements.
type OrderByBuilder struct {
	sql.Builder
	orders []any
}

// OrderBy creates a new ORDER BY statement builder.
func OrderBy(columns ...string) *OrderByBuilder {
	return (&OrderByBuilder{}).OrderBy(columns...)
}

// OrderByExpr creates a new ORDER BY statement builder with expressions.
func OrderByExpr(exprs ...sql.Querier) *OrderByBuilder {
	return (&OrderByBuilder{}).OrderByExpr(exprs...)
}

// OrderBy changes the order expressions to the given columns.
func (o *OrderByBuilder) OrderBy(columns ...string) *OrderByBuilder {
	o.orders = make([]any, len(columns))
	for i := range columns {
		o.orders[i] = columns[i]
	}
	return o
}

// OrderByExpr changes the order expressions to the given expressions.
func (o *OrderByBuilder) OrderByExpr(exprs ...sql.Querier) *OrderByBuilder {
	o.orders = make([]any, len(exprs))
	for i := range exprs {
		o.orders[i] = exprs[i]
	}
	return o
}

// AppendExpr adds more order expressions to the ORDER BY statement.
func (o *OrderByBuilder) AppendExpr(expr ...sql.Querier) *OrderByBuilder {
	for i := range expr {
		o.orders = append(o.orders, expr[i])
	}
	return o
}

// AppendName adds more order columns to the ORDER BY statement.
func (o *OrderByBuilder) AppendName(columns ...string) *OrderByBuilder {
	for i := range columns {
		o.orders = append(o.orders, columns[i])
	}
	return o
}

// Query returns the ORDER BY statement representation.
func (o *OrderByBuilder) Query() (string, []any) {
	o.WriteString("ORDER BY")
	if len(o.orders) > 0 {
		o.Pad()
		for i, order := range o.orders {
			if i > 0 {
				o.Comma()
			}
			switch order := order.(type) {
			case string:
				o.Ident(order)
			case sql.Querier:
				o.Join(order)
			}
		}
	}
	return o.String(), o.GetArgs()
}

func (o *OrderByBuilder) stmt() {}

// LimitBuilder is a builder for LIMIT statements.
type LimitBuilder struct {
	sql.Builder
	count int
}

// Limit creates a new LIMIT statement builder.
func Limit(count int) *LimitBuilder {
	return &LimitBuilder{count: count}
}

// Count sets the limit count.
func (l *LimitBuilder) Count(count int) *LimitBuilder {
	l.count = count
	return l
}

// Query returns the LIMIT statement representation.
func (l *LimitBuilder) Query() (string, []any) {
	l.WriteString("LIMIT ")
	l.WriteString(strconv.Itoa(l.count))
	return l.String(), l.GetArgs()
}

func (l *LimitBuilder) stmt() {}

// OffsetBuilder is a builder for OFFSET statements.
type OffsetBuilder struct {
	sql.Builder
	skip  bool
	count int
}

// Offset creates a new OFFSET statement builder.
func Offset(count int) *OffsetBuilder {
	return &OffsetBuilder{count: count}
}

// Skip creates a new SKIP statement builder.
func Skip(count int) *OffsetBuilder {
	return &OffsetBuilder{skip: true, count: count}
}

// Count sets the offset count.
func (o *OffsetBuilder) Count(count int) *OffsetBuilder {
	o.count = count
	return o
}

// Query returns the OFFSET statement representation.
func (o *OffsetBuilder) Query() (string, []any) {
	if o.skip {
		o.WriteString("SKIP ")
	} else {
		o.WriteString("OFFSET ")
	}
	o.WriteString(strconv.Itoa(o.count))
	return o.String(), o.GetArgs()
}

func (o *OffsetBuilder) stmt() {}

// ForBuilder is a builder for FOR statements.
type ForBuilder struct {
	sql.Builder
	element Var
	array   sql.Querier
	offset  string
}

// Element creates a new FOR statement builder with the given element name.
func Element(name string) *ForBuilder {
	return &ForBuilder{element: Var(name)}
}

// Element sets the element variable name.
func (f *ForBuilder) Element(name string) *ForBuilder {
	f.element = Var(name)
	return f
}

// In sets the array expression to iterate over.
func (f *ForBuilder) In(array sql.Querier) *ForBuilder {
	f.array = array
	return f
}

// WithOffset adds WITH OFFSET clause.
func (f *ForBuilder) WithOffset() *ForBuilder {
	f.offset = "offset"
	return f
}

// WithOffsetAs adds WITH OFFSET AS <name> clause.
func (f *ForBuilder) WithOffsetAs(name string) *ForBuilder {
	f.offset = name
	return f
}

// Query returns the FOR statement representation.
func (f *ForBuilder) Query() (string, []any) {
	f.WriteString("FOR")
	if f.element != "" {
		f.Pad()
		f.Join(f.element)
	}
	f.WriteString(" IN")
	if f.array != nil {
		f.Pad()
		f.Join(f.array)
	}
	if f.offset != "" {
		f.WriteString(" WITH OFFSET")
		if f.offset != "" && f.offset != "offset" {
			f.WriteString(" AS ")
			f.Ident(f.offset)
		}
	}
	return f.String(), f.GetArgs()
}

// F returns the field access expression for the element variable.
func (f *ForBuilder) F(fields ...string) string {
	return f.element.F(fields...)
}

// Offset returns the offset variable name.
func (f *ForBuilder) Offset() string {
	if f.offset == "" {
		f.offset = "offset"
	}
	return fmt.Sprintf("`%s`", f.offset)
}

func (f *ForBuilder) stmt() {}

// WithBuilder is a builder for WITH statements.
type WithBuilder struct {
	sql.Builder
	all      bool
	distinct bool
	items    []string
	groupBy  []sql.Querier
}

// With creates a new WITH statement builder.
func With(items ...string) *WithBuilder {
	return (&WithBuilder{}).Append(items...)
}

// WithAll creates a new WITH ALL statement builder.
func WithAll(items ...string) *WithBuilder {
	return (&WithBuilder{all: true, distinct: false}).Append(items...)
}

// WithDistinct creates a new WITH DISTINCT statement builder.
func WithDistinct(items ...string) *WithBuilder {
	return (&WithBuilder{all: false, distinct: true}).Append(items...)
}

// Append adds more return items to the WITH statement.
func (w *WithBuilder) Append(items ...string) *WithBuilder {
	w.items = append(w.items, items...)
	return w
}

// GroupBy adds a GROUP BY clause.
func (w *WithBuilder) GroupBy(exprs ...sql.Querier) *WithBuilder {
	w.groupBy = append(w.groupBy, exprs...)
	return w
}

// Query returns the WITH statement representation.
func (w *WithBuilder) Query() (string, []any) {
	w.WriteString("WITH")
	if w.all {
		w.WriteString(" ALL")
	} else if w.distinct {
		w.WriteString(" DISTINCT")
	}
	if len(w.items) > 0 {
		w.Pad()
		w.WriteString(strings.Join(w.items, ", "))
	}
	if len(w.groupBy) > 0 {
		w.NewLine()
		w.WriteString("GROUP BY ")
		w.JoinComma(w.groupBy...)
	}
	return w.String(), w.GetArgs()
}

func (w *WithBuilder) stmt() {}

// StatementKeyword represents simple SQL statement keywords that implement sql.Querier.
type StatementKeyword string

func (sk StatementKeyword) Query() (string, []any) {
	if sk == Next {
		return "\n" + string(sk) + "\n", nil
	}
	return string(sk), nil
}

func (sk StatementKeyword) stmt() {}

// Set operation constants for composite queries.
const (
	UnionAll          StatementKeyword = "UNION ALL"
	UnionDistinct     StatementKeyword = "UNION DISTINCT"
	IntersectAll      StatementKeyword = "INTERSECT ALL"
	IntersectDistinct StatementKeyword = "INTERSECT DISTINCT"
	ExceptAll         StatementKeyword = "EXCEPT ALL"
	ExceptDistinct    StatementKeyword = "EXCEPT DISTINCT"
	Next              StatementKeyword = "NEXT"
)

// GraphTableWrapper is a builder for GRAPH_TABLE operator in SQL queries.
type GraphTableWrapper struct {
	*GraphQuery
	sql.TableView
	alias string
}

// GraphTable creates a new GRAPH_TABLE operator builder.
func GraphTable(query *GraphQuery) *GraphTableWrapper {
	return &GraphTableWrapper{
		GraphQuery: query,
	}
}

// As sets the alias for the table.
func (g *GraphTableWrapper) As(alias string) *GraphTableWrapper {
	g.alias = alias
	return g
}

// Query returns the GRAPH_TABLE operator representation.
func (g *GraphTableWrapper) Query() (string, []any) {
	b := g.Clone()
	b.WriteString("GRAPH_TABLE ")
	b.Wrap(func(b *sql.Builder) {
		if g.graph != "" {
			b.Indent().NewLine()
			b.Ident(g.graph)
		}

		if len(g.stmts) > 0 {
			b.NewLine()
			b.JoinNewLine(g.stmts...)
			b.Dedent().NewLine()
		}
	})

	if g.alias != "" {
		b.WriteString(" AS ")
		b.Ident(g.alias)
	}

	return b.String(), b.GetArgs()
}

func (g *GraphTableWrapper) C(column string) string {
	if g.IsQualified(column) {
		return column
	}
	name := g.alias
	b := &sql.Builder{}
	b.Ident(name).WriteByte('.').Ident(column)
	return b.String()
}

// Concat concatenates two path patterns with the || operator.
func Concat(p, q *PathPattern) sql.Querier {
	var b sql.Builder
	if p.variable != "" {
		b.Join(p.variable)
	} else {
		b.Wrap(func(b *sql.Builder) {
			b.Join(p)
		})
	}
	b.WriteString(" || ")
	if q.variable != "" {
		b.Join(q.variable)
	} else {
		b.Wrap(func(b *sql.Builder) {
			b.Join(q)
		})
	}
	return sql.Expr(b.String(), b.GetArgs()...)
}

type GraphPatternBuilder struct {
	sql.Builder
	patterns  []sql.Querier
	where     *sql.Predicate
	collected [][]*sql.Predicate
}

func GraphPattern(patterns ...Pattern) *GraphPatternBuilder {
	return (&GraphPatternBuilder{}).AppendPatterns(patterns...)
}

func (g *GraphPatternBuilder) AppendPatterns(patterns ...Pattern) *GraphPatternBuilder {
	for _, p := range patterns {
		g.patterns = append(g.patterns, p)
	}
	return g
}

func (g *GraphPatternBuilder) Where(pred *sql.Predicate) *GraphPatternBuilder {
	if len(g.collected) > 0 {
		g.collected[len(g.collected)-1] = append(g.collected[len(g.collected)-1], pred)
		return g
	}
	if g.where == nil {
		g.where = pred
	} else {
		g.where = sql.And(g.where, pred)
	}
	return g
}

func (w *GraphPatternBuilder) Query() (string, []any) {
	w.JoinComma(w.patterns...)
	w.NewLine()
	if w.where != nil {
		w.WriteString("WHERE ")
		w.Join(w.where)
	}
	return w.String(), w.GetArgs()
}

// CollectPredicates indicates the appended predicates should be collected
// and not appended to the WHERE clause.
func (g *GraphPatternBuilder) CollectPredicates() *GraphPatternBuilder {
	g.collected = append(g.collected, []*sql.Predicate{})
	return g
}

// CollectedPredicates returns the collected predicates.
func (g *GraphPatternBuilder) CollectedPredicates() []*sql.Predicate {
	if len(g.collected) == 0 {
		return nil
	}
	return g.collected[len(g.collected)-1]
}

// UncollectedPredicates stop collecting predicates.
func (g *GraphPatternBuilder) UncollectedPredicates() *GraphPatternBuilder {
	if len(g.collected) > 0 {
		g.collected = g.collected[:len(g.collected)-1]
	}
	return g
}

func (g *GraphPatternBuilder) pattern() {}

// Pattern is an interface for graph patterns.
type Pattern interface {
	sql.Querier
	pattern()
}

// PathPattern is a builder for path patterns.
type PathPattern struct {
	sql.Builder
	variable     Var
	searchPrefix PathSearchPrefix
	pathMode     PathMode
	elements     []sql.Querier
	bound        [2]*int
}

// Path creates a new path pattern builder.
func Path(path *PathPattern) *PathPattern {
	return (&PathPattern{}).Path(path)
}

// From creates a new path pattern builder with a starting node.
func From(node *NodePattern) *PathPattern {
	return (&PathPattern{}).From(node)
}

// To creates a new path pattern builder with an ending node.
func To(node *NodePattern) *PathPattern {
	return (&PathPattern{}).To(node)
}

// Via creates a new path pattern builder with an edge.
func Via(edge *EdgePattern) *PathPattern {
	return (&PathPattern{}).Via(edge)
}

// From adds a starting node to the path pattern.
func (p *PathPattern) From(node *NodePattern) *PathPattern {
	p.elements = append(p.elements, node)
	return p
}

// To adds an ending node to the path pattern.
func (p *PathPattern) To(node *NodePattern) *PathPattern {
	p.elements = append(p.elements, node)
	return p
}

// Via adds an edge to the path pattern.
func (p *PathPattern) Via(edge *EdgePattern) *PathPattern {
	p.elements = append(p.elements, edge)
	return p
}

// Path adds a sub-path to the path pattern.
func (p *PathPattern) Path(path *PathPattern) *PathPattern {
	p.elements = append(p.elements, path)
	return p
}

// Named sets the path variable name.
func (p *PathPattern) Named(name string) *PathPattern {
	p.variable = Var(name)
	return p
}

// SearchPrefix sets the path search prefix.
func (p *PathPattern) SearchPrefix(prefix PathSearchPrefix) *PathPattern {
	p.searchPrefix = prefix
	return p
}

// All sets the path search prefix to ALL.
func (p *PathPattern) All() *PathPattern {
	p.searchPrefix = PrefixAll
	return p
}

// Any sets the path search prefix to ANY.
func (p *PathPattern) Any() *PathPattern {
	p.searchPrefix = PrefixAny
	return p
}

// AnyShortest sets the path search prefix to ANY SHORTEST.
func (p *PathPattern) AnyShortest() *PathPattern {
	p.searchPrefix = PrefixAnyShortest
	return p
}

// Mode sets the path mode.
func (p *PathPattern) Mode(mode PathMode) *PathPattern {
	p.pathMode = mode
	return p
}

// Walk sets the path mode to WALK.
func (p *PathPattern) Walk() *PathPattern {
	p.pathMode = ModeWalk
	return p
}

// Acyclic sets the path mode to ACYCLIC.
func (p *PathPattern) Acyclic() *PathPattern {
	p.pathMode = ModeAcyclic
	return p
}

// Trail sets the path mode to TRAIL.
func (p *PathPattern) Trail() *PathPattern {
	p.pathMode = ModeTrail
	return p
}

// Fixed sets the path quantifier to fixed with the given bound.
func (p *PathPattern) Fixed(bound int) *PathPattern {
	p.bound[0] = &bound
	p.bound[1] = nil
	return p
}

// Bounded sets the path quantifier to bounded with the given lower and upper bounds.
func (p *PathPattern) Bounded(lower, upper *int) *PathPattern {
	p.bound[0] = lower
	p.bound[1] = upper
	return p
}

// F returns the field access expression for the path variable.
func (p *PathPattern) F(fields ...string) string {
	return p.variable.F(fields...)
}

// Query returns the path pattern representation.
func (p *PathPattern) Query() (string, []any) {
	b := p.Clone()
	quantified := p.bound[0] != nil || p.bound[1] != nil
	if quantified {
		b.WriteByte('(')
	}
	if p.variable != "" {
		b.Join(p.variable)
		b.WriteString(" = ")
	}
	if p.searchPrefix != PrefixAll {
		b.WriteString(pathSearchPrefixes[p.searchPrefix]).Pad()
	}
	if p.pathMode != ModeWalk && p.searchPrefix == PrefixAll {
		b.WriteString(pathModes[p.pathMode]).Pad()
	}
	b.Join(p.elements...)
	if quantified {
		b.WriteByte(')')
	}
	if quantified {
		b.WrapBraces(func(b *sql.Builder) {
			if p.bound[0] != nil {
				b.WriteString(strconv.Itoa(*p.bound[0]))
			}
			if p.bound[1] != nil {
				b.Comma()
				b.WriteString(strconv.Itoa(*p.bound[1]))
			}
		})
	}
	return b.String(), b.GetArgs()
}

func (p *PathPattern) pattern() {}

// NodePattern is a builder for node patterns.
type NodePattern struct {
	sql.Builder
	filler    *patternFiller
	collected [][]*sql.Predicate
}

// N creates a new node pattern builder.
func N() *NodePattern {
	return &NodePattern{
		filler: &patternFiller{properties: make(map[string]sql.Querier)},
	}
}

// NodeL creates a new node pattern builder with the given label.
func NodeL(label string) *NodePattern {
	return N().Labels(label)
}

// Node creates a new node pattern builder.
func Node(name string, labels ...string) *NodePattern {
	return N().Named(name).Labels(labels...)
}

// Named sets the node variable.
func (n *NodePattern) Named(name string) *NodePattern {
	n.filler.variable = Var(name)
	return n
}

// Labels sets the node labels using simple OR logic.
func (n *NodePattern) Labels(labels ...string) *NodePattern {
	ls := make([]*labelExpr, len(labels))
	for i, l := range labels {
		ls[i] = L(l)
	}
	n.filler.labelExpr = OrL(ls...)
	return n
}

// LabelExpression sets a complex label expression.
func (n *NodePattern) LabelExpr(expr *labelExpr) *NodePattern {
	n.filler.labelExpr = expr
	return n
}

func (n *NodePattern) Property(key string, value any) *NodePattern {
	n.filler.properties[key] = sql.V(value)
	return n
}

func (n *NodePattern) Properties(props map[string]any) *NodePattern {
	for k, v := range props {
		n.filler.properties[k] = sql.V(v)
	}
	return n
}

// PropertyExpr adds a property filter.
func (n *NodePattern) PropertyExpr(key string, expr sql.Querier) *NodePattern {
	n.filler.properties[key] = expr
	return n
}

// PropertiesExpr adds multiple property filters.
func (n *NodePattern) PropertiesExpr(props map[string]sql.Querier) *NodePattern {
	maps.Copy(n.filler.properties, props)
	return n
}

// Where adds a WHERE condition.
func (n *NodePattern) Where(condition *sql.Predicate) *NodePattern {
	if len(n.collected) > 0 {
		n.collected[len(n.collected)-1] = append(n.collected[len(n.collected)-1], condition)
		return n
	}
	if n.filler.where == nil {
		n.filler.where = condition
	} else {
		n.filler.where = sql.And(n.filler.where, condition)
	}
	return n
}

// F returns the field access expression for the node variable.
func (n *NodePattern) F(fields ...string) string {
	var b sql.Builder
	if n.filler.variable != "" {
		b.Join(n.filler.variable)
	}
	for _, field := range fields {
		b.WriteByte('.').Ident(field)
	}
	return b.String()
}

// L returns a label expression for the node pattern.
func (n *NodePattern) L() *labelExpr {
	return n.filler.labelExpr
}

// Query returns the node pattern representation.
func (n *NodePattern) Query() (string, []any) {
	b := n.Clone()
	b.Wrap(func(b *sql.Builder) {
		b.Join(n.filler)
	})
	return b.String(), b.GetArgs()
}

// CollectPredicates indicates the appended predicates should be collected
// and not appended to the WHERE clause.
func (n *NodePattern) CollectPredicates() *NodePattern {
	n.collected = append(n.collected, []*sql.Predicate{})
	return n
}

// CollectedPredicates returns the collected predicates.
func (n *NodePattern) CollectedPredicates() []*sql.Predicate {
	if len(n.collected) == 0 {
		return nil
	}
	return n.collected[len(n.collected)-1]
}

// UncollectedPredicates stop collecting predicates.
func (n *NodePattern) UncollectedPredicates() *NodePattern {
	if len(n.collected) > 0 {
		n.collected = n.collected[:len(n.collected)-1]
	}
	return n
}

func (n *NodePattern) pattern() {}

// EdgeDirection represents edge direction types.
type EdgeDirection int

const (
	EdgeAnyDirection EdgeDirection = iota
	EdgeLeft
	EdgeRight
)

// EdgePattern is a builder for edge patterns.
type EdgePattern struct {
	sql.Builder
	filler      *patternFiller
	direction   EdgeDirection
	abbreviated bool
	collected   [][]*sql.Predicate
}

// E creates a new edge pattern builder.
func E() *EdgePattern {
	return &EdgePattern{
		direction: EdgeAnyDirection,
		filler:    &patternFiller{properties: make(map[string]sql.Querier)},
	}
}

// EdgeL creates a new edge pattern builder with the given label.
func EdgeL(label string) *EdgePattern {
	return E().Labels(label)
}

// Edge creates a new edge pattern builder.
func Edge(name string, labels ...string) *EdgePattern {
	return E().Named(name).Labels(labels...)
}

// In sets the edge direction to left.
func (e *EdgePattern) In() *EdgePattern {
	e.direction = EdgeLeft
	return e
}

// Out sets the edge direction to right.
func (e *EdgePattern) Out() *EdgePattern {
	e.direction = EdgeRight
	return e
}

// Any sets the edge direction to any.
func (e *EdgePattern) Any() *EdgePattern {
	e.direction = EdgeAnyDirection
	return e
}

// Named sets the edge variable.
func (e *EdgePattern) Named(name string) *EdgePattern {
	e.filler.variable = Var(name)
	return e
}

// Labels sets the edge labels using simple OR logic.
func (e *EdgePattern) Labels(labels ...string) *EdgePattern {
	ls := make([]*labelExpr, len(labels))
	for i, l := range labels {
		ls[i] = L(l)
	}
	e.filler.labelExpr = OrL(ls...)
	return e
}

// LabelExpr sets a complex label expression.
func (e *EdgePattern) LabelExpr(expr *labelExpr) *EdgePattern {
	e.filler.labelExpr = expr
	return e
}

// Property adds a property filter.
func (e *EdgePattern) Property(key string, value any) *EdgePattern {
	e.filler.properties[key] = sql.V(value)
	return e
}

// Properties adds multiple property filters.
func (e *EdgePattern) Properties(props map[string]any) *EdgePattern {
	for k, v := range props {
		e.filler.properties[k] = sql.V(v)
	}
	return e
}

// PropertyExpr adds a property filter.
func (e *EdgePattern) PropertyExpr(key string, value sql.Querier) *EdgePattern {
	e.filler.properties[key] = value
	return e
}

// Properties adds multiple property filters.
func (e *EdgePattern) propertiesExpr(props map[string]sql.Querier) *EdgePattern {
	maps.Copy(e.filler.properties, props)
	return e
}

// Where adds a WHERE condition.
func (e *EdgePattern) Where(condition *sql.Predicate) *EdgePattern {
	if len(e.collected) > 0 {
		e.collected[len(e.collected)-1] = append(e.collected[len(e.collected)-1], condition)
		return e
	}
	if e.filler.where == nil {
		e.filler.where = condition
	} else {
		e.filler.where = sql.And(e.filler.where, condition)
	}
	return e
}

// Abbreviated makes this an abbreviated edge pattern.
func (e *EdgePattern) Abbreviated() *EdgePattern {
	e.abbreviated = true
	return e
}

// F returns the field access expression for the edge variable.
func (e *EdgePattern) F(fields ...string) string {
	return e.filler.variable.F(fields...)
}

// Query returns the edge pattern representation.
func (e *EdgePattern) Query() (string, []any) {
	b := e.Clone()
	if e.direction == EdgeLeft {
		b.WriteByte('<')
	}
	if !e.abbreviated {
		b.WriteString("-[")
		b.Join(e.filler)
		b.WriteString("]-")
	} else {
		b.WriteByte('-')
	}
	if e.direction == EdgeRight {
		b.WriteByte('>')
	}
	return b.String(), b.GetArgs()
}

// CollectPredicates indicates the appended predicates should be collected
// and not appended to the WHERE clause.
func (e *EdgePattern) CollectPredicates() *EdgePattern {
	e.collected = append(e.collected, []*sql.Predicate{})
	return e
}

// CollectedPredicates returns the collected predicates.
func (e *EdgePattern) CollectedPredicates() []*sql.Predicate {
	if len(e.collected) == 0 {
		return nil
	}
	return e.collected[len(e.collected)-1]
}

// UncollectedPredicates stop collecting predicates.
func (e *EdgePattern) UncollectedPredicates() *EdgePattern {
	if len(e.collected) > 0 {
		e.collected = e.collected[:len(e.collected)-1]
	}
	return e
}

func (e *EdgePattern) pattern() {}

type patternFiller struct {
	sql.Builder
	variable   Var
	labelExpr  *labelExpr
	properties map[string]sql.Querier
	where      *sql.Predicate
}

func (p *patternFiller) Query() (string, []any) {
	b := p.Clone()
	if p.variable != "" {
		b.Join(p.variable)
	}
	if p.labelExpr != nil {
		b.WriteByte(':')
		b.Join(p.labelExpr)
	}
	if len(p.properties) > 0 {
		b.Pad()
		b.WrapBraces(func(b *sql.Builder) {
			first := true
			for key, value := range p.properties {
				if !first {
					b.Comma()
				}
				b.WriteString(key)
				b.WriteString(": ")
				b.Join(value)
				first = false
			}
		})
	}
	if p.where != nil {
		b.WriteString(" WHERE ")
		b.Join(p.where)
	}
	return b.String(), b.GetArgs()
}

// labelExpr is a builder for complex label expressions.
type labelExpr struct {
	sql.Builder
	depth int
	fns   []func(*sql.Builder)
}

// L creates a new label expression with the given label name.
func L(name string) *labelExpr {
	return &labelExpr{
		fns: []func(*sql.Builder){
			func(b *sql.Builder) {
				if name == "" {
					b.WriteByte('%')
					return
				}
				b.Ident(name)
			},
		},
	}
}

func AndL(labels ...*labelExpr) *labelExpr {
	l := &labelExpr{}
	return l.Append(func(b *sql.Builder) {
		l.mayWrap(labels, b, '&')
	})
}

func OrL(labels ...*labelExpr) *labelExpr {
	l := &labelExpr{}
	return l.Append(func(b *sql.Builder) {
		l.mayWrap(labels, b, '|')
	})
}

func NotL(label *labelExpr) *labelExpr {
	l := &labelExpr{}
	return l.Append(func(b *sql.Builder) {
		b.WriteByte('!')
		if len(label.fns) > 1 && l.depth != 0 {
			b.WriteByte('(')
			b.Join(label)
			b.WriteByte(')')
		} else {
			b.Join(label)
		}
	})
}

// Append adds a function to the label expression builder.
func (l *labelExpr) Append(fns ...func(*sql.Builder)) *labelExpr {
	l.fns = append(l.fns, fns...)
	return l
}

func (l *labelExpr) mayWrap(exprs []*labelExpr, b *sql.Builder, op byte) {
	switch n := len(exprs); {
	case n == 1:
		b.Join(exprs[0])
		return
	case n > 1 && l.depth != 0:
		b.WriteByte('(')
		defer b.WriteByte(')')
	}
	for i := range exprs {
		exprs[i].depth = l.depth + 1
		if i > 0 {
			b.WriteByte(op)
		}
		if len(exprs[i].fns) > 1 {
			b.Wrap(func(b *sql.Builder) {
				b.Join(exprs[i])
			})
		} else {
			b.Join(exprs[i])
		}
	}
}

// String returns the label expression representation.
func (l *labelExpr) Query() (string, []any) {
	if l.Len() > 0 || len(l.GetArgs()) > 0 {
		l.Reset()
		l.ClearArgs()
	}
	for _, f := range l.fns {
		f(&l.Builder)
	}
	return l.String(), l.GetArgs()
}

// PathSearchPrefix represents path search prefix types.
type PathSearchPrefix int

const (
	PrefixAll PathSearchPrefix = iota
	PrefixAny
	PrefixAnyShortest
)

var pathSearchPrefixes = [...]string{
	PrefixAll:         "ALL",
	PrefixAny:         "ANY",
	PrefixAnyShortest: "ANY SHORTEST",
}

// PathMode represents path mode types.
type PathMode int

const (
	ModeWalk PathMode = iota
	ModeAcyclic
	ModeTrail
)

var pathModes = [...]string{
	ModeWalk:    "WALK",
	ModeAcyclic: "ACYCLIC",
	ModeTrail:   "TRAIL",
}

// InQuery returns the `IN` subquery predicate for the given expression.
func InQuery(value any, expr *GraphQuery) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Arg(value)
		b.WriteString(" IN ")
		b.WrapBraces(func(b *sql.Builder) {
			b.Indent().NewLine()
			b.Join(expr)
			b.Dedent().NewLine()
		})
	})
}

// NotInQuery returns the `NOT IN` subquery predicate for the given expression.
func NotInQuery(value any, expr *GraphQuery) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Arg(value)
		b.WriteString(" NOT IN ")
		b.WrapBraces(func(b *sql.Builder) {
			b.Indent().NewLine()
			b.Join(expr)
			b.Dedent().NewLine()
		})
	})
}

// ExistsQuery returns an `EXISTS` subquery predicate for the given expression.
func ExistsQuery[T *GraphQuery | *Matcher | *GraphPatternBuilder | *PathPattern | *NodePattern | *EdgePattern](expr T) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("EXISTS ")
		b.WrapBraces(func(b *sql.Builder) {
			b.Indent().NewLine()
			b.Join(sql.Querier(expr))
			b.Dedent().NewLine()
		})
	})
}

// IsLabeled returns an "IS LABELED" predicate.
func IsLabeled(element string, labelExpr *labelExpr) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Ident(element)
		b.WriteString(" IS LABELED ")
		if labelExpr != nil {
			b.Join(labelExpr)
		}
	})
}

// IsNotLabeled returns an "IS NOT LABELED" predicate.
func IsNotLabeled(element string, labelExpr *labelExpr) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Ident(element)
		b.WriteString(" IS NOT LABELED ")
		if labelExpr != nil {
			b.Join(labelExpr)
		}
	})
}

// IsSource returns an "IS SOURCE" predicate.
func IsSource(node, edge string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Ident(node)
		b.WriteString(" IS SOURCE OF ")
		b.Ident(edge)
	})
}

// IsNotSource returns an "IS NOT SOURCE" predicate.
func IsNotSource(node, edge string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Ident(node)
		b.WriteString(" IS NOT SOURCE OF ")
		b.Ident(edge)
	})
}

// IsDestination returns an "IS DESTINATION" predicate.
func IsDestination(node, edge string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Ident(node)
		b.WriteString(" IS DESTINATION OF ")
		b.Ident(edge)
	})
}

// IsNotDestination returns an "IS NOT DESTINATION" predicate.
func IsNotDestination(node, edge string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Ident(node)
		b.WriteString(" IS NOT DESTINATION OF ")
		b.Ident(edge)
	})
}

// AllDifferent returns an "ALL_DIFFERENT" predicate.
func AllDifferent(elements ...string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("ALL_DIFFERENT")
		b.Wrap(func(b *sql.Builder) {
			b.WriteString(strings.Join(elements, ", "))
		})
	})
}

// Same returns a "SAME" predicate.
func Same(elements ...string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("SAME")
		b.Wrap(func(b *sql.Builder) {
			b.WriteString(strings.Join(elements, ", "))
		})
	})
}

// PropertyExists returns a "PROPERTY_EXISTS" predicate.
func PropertyExists(element, property string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("PROPERTY_EXISTS(")
		b.WriteString(element)
		b.Comma()
		b.WriteString(property)
		b.WriteByte(')')
	})
}

// DestinationNodeID returns a new DESTINATION_NODE_ID function.
func DestinationNodeID(edge string) *sql.Func {
	f := &sql.Func{}
	f.ByName("DESTINATION_NODE_ID", edge)
	return f
}

// Edges returns a new EDGES function.
func Edges(path string) *sql.Func {
	f := &sql.Func{}
	f.ByName("EDGES", path)
	return f
}

// ElementID returns a new ELEMENT_ID function.
func ElementID(element string) *sql.Func {
	f := &sql.Func{}
	f.ByName("ELEMENT_ID", element)
	return f
}

// IsAcyclic returns a new IS_ACYCLIC predicate.
func IsAcyclic(path string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("IS_ACYCLIC(")
		b.WriteString(path)
		b.WriteByte(')')
	})
}

// IsSimple returns a new IS_SIMPLE predicate.
func IsSimple(path string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("IS_SIMPLE(")
		b.WriteString(path)
		b.WriteByte(')')
	})
}

// IsTrail returns a new IS_TRAIL predicate.
func IsTrail(path string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("IS_TRAIL(")
		b.WriteString(path)
		b.WriteByte(')')
	})
}

// LabelsFunc returns a new LABELS function.
func LabelsFunc(element string) *sql.Func {
	f := &sql.Func{}
	f.ByName("LABELS", element)
	return f
}

// Nodes returns a new NODES function.
func Nodes(path string) *sql.Func {
	f := &sql.Func{}
	f.ByName("NODES", path)
	return f
}

// Paths returns a new PATH function.
func Paths(elements ...string) *sql.Func {
	f := &sql.Func{}
	f.ByName("PATH", elements...)
	return f
}

// PathFirst returns a new PATH_FIRST function.
func PathFirst(path string) *sql.Func {
	f := &sql.Func{}
	f.ByName("PATH_FIRST", path)
	return f
}

// PathLast returns a new PATH_LAST function.
func PathLast(path string) *sql.Func {
	f := &sql.Func{}
	f.ByName("PATH_LAST", path)
	return f
}

// PathLength returns a new PATH_LENGTH function.
func PathLength(path string) *sql.Func {
	f := &sql.Func{}
	f.ByName("PATH_LENGTH", path)
	return f
}

// PropertyNames returns a new PROPERTY_NAMES function.
func PropertyNames(element string) *sql.Func {
	f := &sql.Func{}
	f.ByName("PROPERTY_NAMES", element)
	return f
}

// SourceNodeID returns a new SOURCE_NODE_ID function.
func SourceNodeID(edge string) *sql.Func {
	f := &sql.Func{}
	f.ByName("SOURCE_NODE_ID", edge)
	return f
}

// ArrayQuery returns an ARRAY expression for the given subquery.
func ArrayQuery(expr *GraphQuery) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		b.WriteString("ARRAY ")
		b.WrapBraces(func(b *sql.Builder) {
			b.Indent().NewLine()
			b.Join(expr)
			b.Dedent().NewLine()
		})
	})
}

// ValueQuery returns a VALUE expression for the given subquery.
func ValueQuery(expr *GraphQuery) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		b.WriteString("VALUE ")
		b.WrapBraces(func(b *sql.Builder) {
			b.Indent().NewLine()
			b.Join(expr)
			b.Dedent().NewLine()
		})
	})
}

type Var string

func (v Var) Query() (string, []any) {
	var b sql.Builder
	if b.IsIdent(string(v)) {
		b.WriteString(string(v))
	} else {
		b.Ident(string(v))
	}
	return b.Query()
}

func (v Var) F(fields ...string) string {
	var b sql.Builder
	b.Ident(string(v))
	for _, f := range fields {
		if strings.HasPrefix(f, "[") && strings.HasSuffix(f, "]") {
			b.WriteString(f)
		} else {
			b.WriteByte('.')
			b.Ident(f)
		}
	}
	return b.String()
}
