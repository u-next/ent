package sqlpgq

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"entgo.io/ent/dialect/sql"
)

// Statement is a linear query statement.
type Statement interface {
	sql.Querier
	stmt()
}

// GraphQuery is a builder for complete GQL queries.
type GraphQuery struct {
	sql.Builder
	graph string      // property graph name
	stmts []Statement // linear query statements
}

// Graph returns a new GraphQuery for the GRAPH statement.
func Graph(name string) *GraphQuery {
	return &GraphQuery{
		graph: name,
	}
}

// Graph sets the property graph name.
func (g *GraphQuery) Graph(name string) *GraphQuery {
	g.graph = name
	return g
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
func (g *GraphQuery) MatchWithHint(hint Hint, patterns ...Pattern) *GraphQuery {
	return g.Append(Match().Hint(hint).Append(patterns...))
}

// OptionalMatch adds an OPTIONAL MATCH statement to the linear query.
func (g *GraphQuery) OptionalMatch(patterns ...Pattern) *GraphQuery {
	return g.Append(OptionalMatch().Append(patterns...))
}

// OptionalMatchWithHint adds an OPTIONAL MATCH statement with a hint to the linear query.
func (g *GraphQuery) OptionalMatchWithHint(hint Hint, patterns ...Pattern) *GraphQuery {
	return g.Append(OptionalMatch().Hint(hint).Append(patterns...))
}

// Filter adds a FILTER statement to the linear query.
func (g *GraphQuery) Filter(p *Predicate) *GraphQuery {
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
func (g *GraphQuery) Return(items ...*columnExpr) *GraphQuery {
	return g.Append(Return(items...))
}

// ReturnAll adds a RETURN ALL statement to the linear query.
func (g *GraphQuery) ReturnAll(items ...*columnExpr) *GraphQuery {
	return g.Append(ReturnAll(items...))
}

// ReturnDistinct adds a RETURN DISTINCT statement to the linear query.
func (g *GraphQuery) ReturnDistinct(items ...*columnExpr) *GraphQuery {
	return g.Append(ReturnDistinct(items...))
}

// GroupBy adds a GROUP BY clause to the last RETURN or WITH statement.
func (g *GraphQuery) GroupBy(items ...*columnExpr) *GraphQuery {
	return g.Append(GroupBy(items...))
}

// OrderBy adds an ORDER BY statement to the linear query.
func (g *GraphQuery) OrderBy(obs ...*columnExpr) *GraphQuery {
	return g.Append(OrderBy(obs...))
}

// Limit adds a LIMIT statement to the linear query.
func (g *GraphQuery) Limit(count int64) *GraphQuery {
	return g.Append(Limit(count))
}

// Offset adds an OFFSET statement to the linear query.
func (g *GraphQuery) Offset(count int64) *GraphQuery {
	return g.Append(Offset(count))
}

// Skip is an alias for Offset that uses the SKIP keyword.
func (g *GraphQuery) Skip(count int64) *GraphQuery {
	return g.Append(Skip(count))
}

// With adds a WITH statement to the linear query.
func (g *GraphQuery) With(items ...*columnExpr) *GraphQuery {
	return g.Append(With(items...))
}

// WithAll adds a WITH ALL statement to the linear query.
func (g *GraphQuery) WithAll(items ...*columnExpr) *GraphQuery {
	return g.Append(WithAll(items...))
}

// WithDistinct adds a WITH DISTINCT statement to the linear query.
func (g *GraphQuery) WithDistinct(items ...*columnExpr) *GraphQuery {
	return g.Append(WithDistinct(items...))
}

// Next chains multiple linear queries with NEXT statements.
func (g *GraphQuery) Next() *GraphQuery {
	return g.Append(Next())
}

// UnionAll adds a UNION ALL operation to the composite query.
func (g *GraphQuery) UnionAll() *GraphQuery {
	return g.Append(UnionAll())
}

// UnionDistinct adds a UNION DISTINCT operation to the composite query.
func (g *GraphQuery) UnionDistinct() *GraphQuery {
	return g.Append(UnionDistinct())
}

// IntersectAll adds an INTERSECT ALL operation to the composite query.
func (g *GraphQuery) IntersectAll() *GraphQuery {
	return g.Append(IntersectAll())
}

// IntersectDistinct adds an INTERSECT DISTINCT operation to the composite query.
func (g *GraphQuery) IntersectDistinct() *GraphQuery {
	return g.Append(IntersectDistinct())
}

// ExceptAll adds an EXCEPT ALL operation to the composite query.
func (g *GraphQuery) ExceptAll() *GraphQuery {
	return g.Append(ExceptAll())
}

// ExceptDistinct adds an EXCEPT DISTINCT operation to the composite query.
func (g *GraphQuery) ExceptDistinct() *GraphQuery {
	return g.Append(ExceptDistinct())
}

// Query returns the GQL query representation.
func (g *GraphQuery) Query() (string, []any) {
	if g.graph != "" {
		g.WriteString("GRAPH ").Ident(g.graph).NewLine()
	}

	for i, stmt := range g.stmts {
		if i > 0 {
			g.NewLine()
		}
		query, args := stmt.Query()
		g.WriteString(query)
		g.Args(args...)
	}
	return g.String(), g.GetArgs()
}

func (g *GraphQuery) Clone() *GraphQuery {
	return &GraphQuery{
		Builder: g.Builder.Clone(),
		graph:   g.graph,
		stmts:   g.stmts,
	}
}

// Hint represents a Graph query hint.
type Hint map[string]string

// Query returns the hint representation.
func (h Hint) Query() (string, []any) {
	var b sql.Builder
	b.WriteString("@")
	b.WrapBraces(func(b *sql.Builder) {
		i := 0
		for k, v := range h {
			if i > 0 {
				b.Comma()
			}
			b.WriteString(fmt.Sprintf("%s=%s", k, v))
			i++
		}
	})
	return b.String(), b.GetArgs()
}

// Matcher is a builder for MATCH statements.
type Matcher struct {
	sql.Builder
	optional bool
	hint     Hint
	patterns []Pattern
}

// Match creates a new MATCH statement builder.
func Match(patterns ...Pattern) *Matcher {
	return (&Matcher{}).Append(patterns...)
}

// OptionalMatch creates a new OPTIONAL MATCH statement builder.
func OptionalMatch(patterns ...Pattern) *Matcher {
	return (&Matcher{optional: true}).Append(patterns...)
}

// Hint sets the match hint.
func (m *Matcher) Hint(hint Hint) *Matcher {
	m.hint = hint
	return m
}

// Append adds more patterns to the MATCH statement.
func (m *Matcher) Append(patterns ...Pattern) *Matcher {
	m.patterns = append(m.patterns, patterns...)
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
		m.Join(m.hint)
	}
	if m.patterns != nil {
		m.Pad()
		for i, p := range m.patterns {
			if i > 0 {
				m.Comma()
			}
			m.Join(p)
		}
	}
	return m.String(), m.GetArgs()
}

func (m *Matcher) stmt() {}

// FilterBuilder is a builder for FILTER statements.
type FilterBuilder struct {
	sql.Builder
	where     bool
	predicate *Predicate
}

// Filter creates a new FILTER statement builder.
func Filter(pred *Predicate) *FilterBuilder {
	return &FilterBuilder{predicate: pred}
}

// FilterWhere creates a new FILTER statement builder with WHERE.
func FilterWhere(pred *Predicate) *FilterBuilder {
	return &FilterBuilder{where: true, predicate: pred}
}

// Predicate sets the boolean predicate expression to filter by.
func (f *FilterBuilder) Predicate(pred *Predicate) *FilterBuilder {
	f.predicate = pred
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

func (f *FilterBuilder) stmt() {}

// columnExpr represents a single column expression.
type columnExpr struct {
	sql.Builder
	expr    sql.Querier
	alias   string
	star    bool
	collate string
	asc     bool
	desc    bool
}

// AllColumns creates a new column expression representing all columns (*).
func AllColumns() *columnExpr {
	return &columnExpr{star: true}
}

// Const creates a new constant column expression.
func Const(v any) *columnExpr {
	return &columnExpr{expr: Expr("?", v)}
}

// Column creates a new column expression.
func Column(expr string, args ...any) *columnExpr {
	return &columnExpr{expr: Expr(expr, args...)}
}

// As sets the alias for the column expression.
func (r *columnExpr) As(alias string) *columnExpr {
	r.alias = alias
	return r
}

// Collate adds a COLLATE specification to the column expression.
func (r *columnExpr) Collate(collate string) *columnExpr {
	r.collate = collate
	return r
}

// Asc sets the order to ascending (default).
func (r *columnExpr) Asc() *columnExpr {
	r.asc = true
	r.desc = false
	return r
}

// Desc sets the order to descending.
func (r *columnExpr) Desc() *columnExpr {
	r.desc = true
	r.asc = false
	return r
}

func (r *columnExpr) Query() (string, []any) {
	if r.star {
		r.WriteByte('*')
	} else if r.expr != nil {
		r.Join(r.expr)
		if r.alias != "" {
			r.WriteString(" AS ")
			r.Ident(r.alias)
		}
		if r.collate != "" {
			r.WriteString(" COLLATE ")
			r.WriteString(r.collate)
		}
		if r.desc {
			r.WriteString(" DESC")
		} else if r.asc {
			r.WriteString(" ASC")
		}
	}
	return r.String(), r.GetArgs()
}

// ReturnBuilder is a builder for RETURN statements.
type ReturnBuilder struct {
	sql.Builder
	all      bool
	distinct bool
	items    []sql.Querier
	groupBy  []sql.Querier
	orderBy  []sql.Querier
	limit    int64
	offset   int64
}

// Return creates a new RETURN statement builder.
func Return(items ...*columnExpr) *ReturnBuilder {
	return (&ReturnBuilder{}).Append(items...)
}

// ReturnAll creates a new RETURN ALL statement builder.
func ReturnAll(items ...*columnExpr) *ReturnBuilder {
	return (&ReturnBuilder{all: true, distinct: false}).Append(items...)
}

// ReturnDistinct creates a new RETURN DISTINCT statement builder.
func ReturnDistinct(items ...*columnExpr) *ReturnBuilder {
	return (&ReturnBuilder{all: false, distinct: true}).Append(items...)
}

// Append adds more return items to the RETURN statement.
func (r *ReturnBuilder) Append(items ...*columnExpr) *ReturnBuilder {
	for _, column := range items {
		r.items = append(r.items, column)
	}
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
	r.WriteString("RETURN")

	if r.all {
		r.WriteString(" ALL")
	} else if r.distinct {
		r.WriteString(" DISTINCT")
	}

	r.Pad()
	r.JoinComma(r.items...)
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
	variable string
	value    sql.Querier
}

// Assign creates a new assignment expression.
func Assign(name string, value sql.Querier) *assignment {
	return &assignment{
		variable: name,
		value:    value,
	}
}

// Var sets the variable name to assign to.
func (a *assignment) Var(name string) *assignment {
	a.variable = name
	return a
}

// Value sets the expression to assign.
func (a *assignment) Value(value sql.Querier) *assignment {
	a.value = value
	return a
}

// F returns the field name of the variable.
func (a *assignment) F(fields ...string) string {
	var b sql.Builder
	b.Ident(a.variable)
	for _, f := range fields {
		b.WriteString(".")
		b.Ident(f)
	}
	return b.String()
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

// Append adds more assignments to the LET statement.
func (l *LetBuilder) Append(as ...*assignment) *LetBuilder {
	l.assignments = append(l.assignments, as...)
	return l
}

// Query returns the LET statement representation.
func (l *LetBuilder) Query() (string, []any) {
	l.WriteString("LET ")
	for i, assign := range l.assignments {
		if i > 0 {
			l.Comma()
		}
		l.Ident(assign.variable)
		l.WriteString(" = ")
		l.Join(assign.value)
	}
	return l.String(), l.GetArgs()
}

func (l *LetBuilder) stmt() {}

// GroupByBuilder is a builder for GROUP BY statements.
type GroupByBuilder struct {
	sql.Builder
	exprs []sql.Querier
}

// GroupBy creates a new GROUP BY statement builder.
func GroupBy(items ...*columnExpr) *GroupByBuilder {
	return (&GroupByBuilder{}).Append(items...)
}

// Append adds more expressions to the GROUP BY statement.
func (g *GroupByBuilder) Append(items ...*columnExpr) *GroupByBuilder {
	for _, item := range items {
		g.exprs = append(g.exprs, item)
	}
	return g
}

// Query returns the GROUP BY statement representation.
func (g *GroupByBuilder) Query() (string, []any) {
	g.WriteString("GROUP BY ")
	g.JoinComma(g.exprs...)
	return g.String(), g.GetArgs()
}

func (g *GroupByBuilder) stmt() {}

// OrderByBuilder is a builder for ORDER BY statements.
type OrderByBuilder struct {
	sql.Builder
	orders []sql.Querier
}

// OrderBy creates a new ORDER BY statement builder.
func OrderBy(items ...*columnExpr) *OrderByBuilder {
	return (&OrderByBuilder{}).Append(items...)
}

// Asc adds an ascending order expression.
func (o *OrderByBuilder) Append(orders ...*columnExpr) *OrderByBuilder {
	for _, order := range orders {
		o.orders = append(o.orders, order)
	}
	return o
}

// Query returns the ORDER BY statement representation.
func (o *OrderByBuilder) Query() (string, []any) {
	o.WriteString("ORDER BY ")
	o.JoinComma(o.orders...)
	return o.String(), o.GetArgs()
}

func (o *OrderByBuilder) stmt() {}

// LimitBuilder is a builder for LIMIT statements.
type LimitBuilder struct {
	sql.Builder
	count int64
}

// Limit creates a new LIMIT statement builder.
func Limit(count int64) *LimitBuilder {
	return &LimitBuilder{count: count}
}

// Count sets the limit count.
func (l *LimitBuilder) Count(count int64) *LimitBuilder {
	l.count = count
	return l
}

// Query returns the LIMIT statement representation.
func (l *LimitBuilder) Query() (string, []any) {
	l.WriteString("LIMIT ")
	l.WriteString(strconv.FormatInt(l.count, 10))
	return l.String(), l.GetArgs()
}

func (l *LimitBuilder) stmt() {}

// OffsetBuilder is a builder for OFFSET statements.
type OffsetBuilder struct {
	sql.Builder
	skip  bool
	count int64
}

// Offset creates a new OFFSET statement builder.
func Offset(count int64) *OffsetBuilder {
	return &OffsetBuilder{count: count}
}

// Skip creates a new SKIP statement builder.
func Skip(count int64) *OffsetBuilder {
	return &OffsetBuilder{skip: true, count: count}
}

// Count sets the offset count.
func (o *OffsetBuilder) Count(count int64) *OffsetBuilder {
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
	o.WriteString(strconv.FormatInt(o.count, 10))
	return o.String(), o.GetArgs()
}

func (o *OffsetBuilder) stmt() {}

// ForBuilder is a builder for FOR statements.
type ForBuilder struct {
	sql.Builder
	element string
	array   sql.Querier
	offset  string
}

// For creates a new FOR statement builder with the given element name.
func For(name string) *ForBuilder {
	return &ForBuilder{element: name}
}

// Element sets the element variable name.
func (f *ForBuilder) Element(name string) *ForBuilder {
	f.element = name
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
	f.WriteString("FOR ")
	f.Ident(f.element)
	f.WriteString(" IN ")
	if f.array != nil {
		f.Join(f.array)
	}
	if f.offset != "" {
		f.WriteString(" WITH OFFSET")
		if f.offset != "offset" {
			f.WriteString(" AS ")
			f.Ident(f.offset)
		}
	}
	return f.String(), f.GetArgs()
}

// Elem returns the element variable name.
func (f *ForBuilder) Elem() string {
	return fmt.Sprintf("`%s`", f.element)
}

// Offset returns the offset variable name.
func (f *ForBuilder) Offset() string {
	if f.offset == "" {
		f.offset = "offset"
	}
	return fmt.Sprintf("`%s`", f.offset)
}

func (f *ForBuilder) stmt() {}

// NextBuilder is a builder for NEXT statements.
type NextBuilder struct {
	sql.Builder
}

// Next creates a new NEXT statement builder.
func Next() *NextBuilder {
	return &NextBuilder{}
}

// Query returns the NEXT statement representation.
func (n *NextBuilder) Query() (string, []any) {
	n.NewLine()
	n.WriteString("NEXT")
	n.NewLine()
	return n.String(), n.GetArgs()
}

func (n *NextBuilder) stmt() {}

// WithBuilder is a builder for WITH statements.
type WithBuilder struct {
	sql.Builder
	all      bool
	distinct bool
	items    []sql.Querier
	groupBy  []sql.Querier
}

// With creates a new WITH statement builder.
func With(items ...*columnExpr) *WithBuilder {
	return (&WithBuilder{}).Append(items...)
}

// WithAll creates a new WITH ALL statement builder.
func WithAll(items ...*columnExpr) *WithBuilder {
	return (&WithBuilder{all: true, distinct: false}).Append(items...)
}

// WithDistinct creates a new WITH DISTINCT statement builder.
func WithDistinct(items ...*columnExpr) *WithBuilder {
	return (&WithBuilder{all: false, distinct: true}).Append(items...)
}

// Append adds more return items to the WITH statement.
func (w *WithBuilder) Append(items ...*columnExpr) *WithBuilder {
	for _, item := range items {
		w.items = append(w.items, item)
	}
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

	w.Pad()
	w.JoinComma(w.items...)

	if len(w.groupBy) > 0 {
		w.NewLine()
		w.WriteString("GROUP BY ")
		w.JoinComma(w.groupBy...)
	}

	return w.String(), w.GetArgs()
}

func (w *WithBuilder) stmt() {}

// SetOp represents different set operation types.
type SetOp int

const (
	SetUnionAll SetOp = iota
	SetUnionDistinct
	SetIntersectAll
	SetIntersectDistinct
	SetExceptAll
	SetExceptDistinct
)

var setOpStrings = [...]string{
	SetUnionAll:          "UNION ALL",
	SetUnionDistinct:     "UNION DISTINCT",
	SetIntersectAll:      "INTERSECT ALL",
	SetIntersectDistinct: "INTERSECT DISTINCT",
	SetExceptAll:         "EXCEPT ALL",
	SetExceptDistinct:    "EXCEPT DISTINCT",
}

type SetOperation func(*sql.Builder)

func (s SetOperation) Query() (string, []any) {
	b := &sql.Builder{}
	s(b)
	return b.String(), b.GetArgs()
}

func (s SetOperation) stmt() {}

func UnionAll() SetOperation {
	return func(b *sql.Builder) {
		// b.WriteOp(SetUnionAll)
	}
}

func UnionDistinct() SetOperation {
	return func(b *sql.Builder) {
		// b.WriteOp(SetUnionDistinct)
	}
}

func IntersectAll() SetOperation {
	return func(b *sql.Builder) {
		// b.WriteOp(SetIntersectAll)
	}
}

func IntersectDistinct() SetOperation {
	return func(b *sql.Builder) {
		// b.WriteOp(SetIntersectDistinct)
	}
}

func ExceptAll() SetOperation {
	return func(b *sql.Builder) {
		// b.WriteOp(SetExceptAll)
	}
}

func ExceptDistinct() SetOperation {
	return func(b *sql.Builder) {
		// b.WriteOp(SetExceptDistinct)
	}
}

// Queries are list of queries join with space between them.
type Queries []sql.Querier

// Query returns query representation of Queriers.
func (n Queries) Query() (string, []any) {
	b := &sql.Builder{}
	for i := range n {
		if i > 0 {
			b.Pad()
		}
		query, args := n[i].Query()
		b.WriteString(query)
		b.Args(args...)
	}
	return b.String(), b.GetArgs()
}

// GraphTableWrapper is a builder for GRAPH_TABLE operator in SQL queries.
type GraphTableWrapper struct {
	*GraphQuery
	sql.TableView
	alias string
}

// GraphTable creates a new GRAPH_TABLE operator builder.
func GraphTable(q *GraphQuery) *GraphTableWrapper {
	return &GraphTableWrapper{
		GraphQuery: q,
	}
}

// As sets the alias for the table.
func (g *GraphTableWrapper) As(alias string) *GraphTableWrapper {
	g.alias = alias
	return g
}

// Query returns the GRAPH_TABLE operator representation.
func (g *GraphTableWrapper) Query() (string, []any) {
	g.WriteString("GRAPH_TABLE ").Wrap(func(b *sql.Builder) {
		if g.graph != "" {
			b.NewLine()
			b.Indent(1).Ident(g.graph)
			b.NewLine()
		}

		for i, stmt := range g.stmts {
			query, args := stmt.Query()
			b.Indent(1).WriteString(query)
			b.Args(args...)
			if i >= 0 {
				b.NewLine()
			}
		}
	})

	if g.alias != "" {
		g.WriteString(" AS ")
		g.Ident(g.alias)
	}

	return g.String(), g.GetArgs()
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

// An Op represents an operator in GQL.
type Op int

// GQL operators.
const (
	OpEQ      Op = iota // =
	OpNEQ               // <>
	OpGT                // >
	OpGTE               // >=
	OpLT                // <
	OpLTE               // <=
	OpIn                // IN
	OpNotIn             // NOT IN
	OpLike              // LIKE
	OpIsNull            // IS NULL
	OpNotNull           // IS NOT NULL
	OpAdd               // +
	OpSub               // -
	OpMul               // *
	OpDiv               // / (Quotient)
	OpMod               // % (Reminder)
	OpAnd               // AND
	OpOr                // OR
	OpNot               // NOT
	// GQL-specific operators
	OpConcat           // || (graph path concatenation)
	OpGraphOr          // | (graph logical OR)
	OpGraphAnd         // & (graph logical AND)
	OpGraphNot         // ! (graph logical NOT)
	OpIsLabeled        // IS LABELED
	OpIsNotLabeled     // IS NOT LABELED
	OpIsSource         // IS SOURCE
	OpIsNotSource      // IS NOT SOURCE
	OpIsDestination    // IS DESTINATION
	OpIsNotDestination // IS NOT DESTINATION
)

var gqlOps = [...]string{
	OpEQ:               "=",
	OpNEQ:              "<>",
	OpGT:               ">",
	OpGTE:              ">=",
	OpLT:               "<",
	OpLTE:              "<=",
	OpIn:               "IN",
	OpNotIn:            "NOT IN",
	OpLike:             "LIKE",
	OpIsNull:           "IS NULL",
	OpNotNull:          "IS NOT NULL",
	OpAdd:              "+",
	OpSub:              "-",
	OpMul:              "*",
	OpDiv:              "/",
	OpMod:              "%",
	OpAnd:              "AND",
	OpOr:               "OR",
	OpNot:              "NOT",
	OpConcat:           "||",
	OpGraphOr:          "|",
	OpGraphAnd:         "&",
	OpGraphNot:         "!",
	OpIsLabeled:        "IS LABELED",
	OpIsNotLabeled:     "IS NOT LABELED",
	OpIsSource:         "IS SOURCE",
	OpIsNotSource:      "IS NOT SOURCE",
	OpIsDestination:    "IS DESTINATION",
	OpIsNotDestination: "IS NOT DESTINATION",
}

// isFunc reports if the given string is a function call.
func isFunc(s string) bool {
	return strings.Contains(s, "(") && strings.Contains(s, ")")
}

// isModifier reports if the given string is a GQL modifier.
func isModifier(s string) bool {
	for _, m := range [...]string{
		"DISTINCT", "ALL", "ASC", "DESC",
		"SHORTEST", "ANY", "ALL_DIFFERENT", "SAME",
		"PROPERTY_EXISTS", "LABELS", "PROPERTIES", "TYPE",
		"PATH", "NODES", "EDGES", "LENGTH",
		"IS LABELED", "IS SOURCE", "IS DESTINATION",
		"IS NOT LABELED", "IS NOT SOURCE", "IS NOT DESTINATION",
		"EXISTS", "WHERE", "OPTIONAL",
	} {
		if strings.HasPrefix(strings.ToUpper(s), m) {
			return true
		}
	}
	return false
}

// isAlias reports if the given string contains an alias.
func isAlias(s string) bool {
	return strings.Contains(s, " AS ") || strings.Contains(s, " as ")
}

// Pattern is an interface for graph patterns.
type Pattern interface {
	sql.Querier
	pattern()
}

// GraphPattern is a builder for graph patterns.
type GraphPattern struct {
	sql.Builder
	pathPatterns []sql.Querier
	where        *Predicate
}

// NewGraphPattern creates a new graph pattern builder.
func NewGraphPattern() *GraphPattern {
	return &GraphPattern{}
}

// Append adds a path pattern to the graph pattern.
func (g *GraphPattern) Append(pattern *PathPattern) *GraphPattern {
	g.pathPatterns = append(g.pathPatterns, pattern)
	return g
}

// Where adds a WHERE clause to the graph pattern.
func (g *GraphPattern) Where(condition *Predicate) *GraphPattern {
	g.where = condition
	return g
}

// Query returns the graph pattern representation.
func (g *GraphPattern) Query() (string, []any) {
	if len(g.pathPatterns) > 0 {
		g.JoinComma(g.pathPatterns...)
	}

	if g.where != nil {
		g.WriteString(" WHERE ")
		g.Join(g.where)
	}

	return g.String(), g.GetArgs()
}

func (g *GraphPattern) pattern() {}

// PathPattern is a builder for path patterns.
type PathPattern struct {
	sql.Builder
	variable     string
	searchPrefix PathSearchPrefix
	pathMode     PathMode
	elements     []Pattern
	quantifier   *QuantifierBuilder
}

// Path creates a new path pattern builder.
func Path() *PathPattern {
	return &PathPattern{}
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

// Variable sets the path variable.
func (p *PathPattern) Variable(name string) *PathPattern {
	p.variable = name
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

// Quantifier sets the quantifier for the path pattern.
func (p *PathPattern) Quantifier(quantifier *QuantifierBuilder) *PathPattern {
	p.quantifier = quantifier
	return p
}

// Fixed sets a fixed quantifier {n} for the path pattern.
func (p *PathPattern) Fixed(bound int) *PathPattern {
	p.quantifier = NewFixedQuantifier(bound)
	return p
}

// Bounded sets a bounded quantifier {min,max} for the path pattern.
func (p *PathPattern) Bounded(lowerBound, upperBound int) *PathPattern {
	p.quantifier = NewBoundedQuantifier(lowerBound, upperBound)
	return p
}

// OpenBounded sets an open bounded quantifier {,max} for the path pattern.
func (p *PathPattern) OpenBounded(upperBound int) *PathPattern {
	p.quantifier = NewOpenBoundedQuantifier(upperBound)
	return p
}

// C returns a formatted string for the table column.
func (p *PathPattern) C(column string) string {
	var b sql.Builder
	if p.variable == "" {
		return column
	}
	b.Ident(p.variable).WriteByte('.').Ident(column)
	return b.String()
}

// Query returns the path pattern representation.
func (p *PathPattern) Query() (string, []any) {
	if p.variable != "" {
		p.WriteString(p.variable)
		p.WriteString(" = ")
	}

	// Path search prefix (cannot be combined with path mode)
	if p.searchPrefix != PrefixAll {
		p.WriteString(pathSearchPrefixes[p.searchPrefix]).Pad()
	}

	// Path mode (cannot be combined with path search prefix other than ALL)
	if p.pathMode != ModeWalk && p.searchPrefix == PrefixAll {
		p.WriteString(pathModes[p.pathMode]).Pad()
	}

	for _, element := range p.elements {
		switch element.(type) {
		case *PathPattern:
			p.Wrap(func(b *sql.Builder) {
				b.Join(element)
			})
		default:
			p.Join(element)
		}
	}

	if p.quantifier != nil {
		quantQuery, quantArgs := p.quantifier.Query()
		p.WriteString(quantQuery)
		p.Args(quantArgs...)
	}

	return p.String(), p.GetArgs()
}

func (p *PathPattern) pattern() {}

// NodePattern is a builder for node patterns.
type NodePattern struct {
	sql.Builder
	variable   string
	labelExpr  *LabelExpr
	properties map[string]sql.Querier
	where      *Predicate
}

// N creates a new node pattern builder.
func N() *NodePattern {
	return &NodePattern{
		properties: make(map[string]sql.Querier),
	}
}

// Named sets the node variable.
func (n *NodePattern) Named(name string) *NodePattern {
	n.variable = name
	return n
}

// Labels sets the node labels using simple OR logic.
func (n *NodePattern) Labels(labels ...string) *NodePattern {
	ls := make([]*LabelExpr, len(labels))
	for i, l := range labels {
		ls[i] = L(l)
	}
	n.labelExpr = OrL(ls...)
	return n
}

// LabelExpression sets a complex label expression.
func (n *NodePattern) LabelExpr(expr *LabelExpr) *NodePattern {
	n.labelExpr = expr
	return n
}

// Property adds a property filter.
func (n *NodePattern) Property(key string, expr sql.Querier) *NodePattern {
	n.properties[key] = expr
	return n
}

// Properties adds multiple property filters.
func (n *NodePattern) Properties(props map[string]sql.Querier) *NodePattern {
	maps.Copy(n.properties, props)
	return n
}

// Where adds a WHERE condition.
func (n *NodePattern) Where(condition *Predicate) *NodePattern {
	n.where = condition
	return n
}

// F returns a formatted string for the field of the node variable.
func (n *NodePattern) F(fields ...string) string {
	var b sql.Builder
	if n.variable != "" {
		b.Ident(n.variable)
	}
	for _, field := range fields {
		b.WriteByte('.').Ident(field)
	}
	return b.String()
}

// L returns a label expression for the node pattern.
func (n *NodePattern) L() *LabelExpr {
	return n.labelExpr
}

// Query returns the node pattern representation.
func (n *NodePattern) Query() (string, []any) {
	if n.Len() > 0 || len(n.GetArgs()) > 0 {
		n.Reset()
		n.ClearArgs()
	}

	n.WriteByte('(')

	if n.variable != "" {
		n.Ident(n.variable)
	}

	if n.labelExpr != nil {
		n.WriteByte(':')
		n.Join(n.labelExpr)
	}

	if len(n.properties) > 0 {
		n.WriteString(" {")
		first := true
		for key, value := range n.properties {
			if !first {
				n.WriteString(", ")
			}
			n.WriteString(key).WriteString(": ")
			n.Join(value)
			first = false
		}
		n.WriteByte('}')
	}

	if n.where != nil {
		n.WriteString(" WHERE ")
		n.Join(n.where)
	}

	n.WriteByte(')')
	return n.String(), n.GetArgs()
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
	direction   EdgeDirection
	variable    string
	labelExpr   *LabelExpr
	properties  map[string]any
	where       sql.Querier
	quantifier  sql.Querier
	abbreviated bool
}

// E creates a new edge pattern builder.
func E() *EdgePattern {
	return &EdgePattern{
		direction:  EdgeAnyDirection,
		properties: make(map[string]any),
	}
}

// LeftDirection sets the edge direction to left.
func (e *EdgePattern) LeftDirection() *EdgePattern {
	e.direction = EdgeLeft
	return e
}

// RightDirection sets the edge direction to right.
func (e *EdgePattern) RightDirection() *EdgePattern {
	e.direction = EdgeRight
	return e
}

// AnyDirection sets the edge direction to any.
func (e *EdgePattern) AnyDirection() *EdgePattern {
	e.direction = EdgeAnyDirection
	return e
}

// Named sets the edge variable.
func (e *EdgePattern) Named(name string) *EdgePattern {
	e.variable = name
	return e
}

// Labels sets the edge labels using simple OR logic.
func (e *EdgePattern) Labels(labels ...string) *EdgePattern {
	ls := make([]*LabelExpr, len(labels))
	for i, l := range labels {
		ls[i] = L(l)
	}
	e.labelExpr = OrL(ls...)
	return e
}

// LabelExpr sets a complex label expression.
func (e *EdgePattern) LabelExpr(expr *LabelExpr) *EdgePattern {
	e.labelExpr = expr
	return e
}

// Property adds a property filter.
func (e *EdgePattern) Property(key string, value any) *EdgePattern {
	e.properties[key] = value
	return e
}

// Properties adds multiple property filters.
func (e *EdgePattern) Properties(props map[string]any) *EdgePattern {
	maps.Copy(e.properties, props)
	return e
}

// Where adds a WHERE condition.
func (e *EdgePattern) Where(condition sql.Querier) *EdgePattern {
	e.where = condition
	return e
}

// Quantifier sets the path quantifier.
func (e *EdgePattern) Quantifier(quantifier sql.Querier) *EdgePattern {
	e.quantifier = quantifier
	return e
}

// Abbreviated makes this an abbreviated edge pattern.
func (e *EdgePattern) Abbreviated() *EdgePattern {
	e.abbreviated = true
	return e
}

// F returns a formatted string for the field of the edge variable.
func (e *EdgePattern) F(fields ...string) string {
	var b sql.Builder
	if e.variable != "" {
		b.Ident(e.variable)
	}
	for _, field := range fields {
		b.WriteByte('.').Ident(field)
	}
	return b.String()
}

// Query returns the edge pattern representation.
func (e *EdgePattern) Query() (string, []any) {
	if e.Len() > 0 || len(e.GetArgs()) > 0 {
		e.Reset()
		e.ClearArgs()
	}

	// Left arrow for left direction
	if e.direction == EdgeLeft {
		e.WriteByte('<')
	}

	// Edge bracket start
	e.WriteByte('-')

	if !e.abbreviated {
		e.WriteByte('[')

		if e.variable != "" {
			e.Ident(e.variable)
		}

		if e.labelExpr != nil {
			e.WriteByte(':')
			e.Join(e.labelExpr)
		}

		if len(e.properties) > 0 {
			e.WriteString(" {")
			first := true
			for key, value := range e.properties {
				if !first {
					e.WriteString(", ")
				}
				e.WriteString(key).WriteString(": ")
				e.Arg(value)
				first = false
			}
			e.WriteByte('}')
		}

		if e.where != nil {
			e.WriteString(" WHERE ")
			e.Join(e.where)
		}

		e.WriteByte(']')
	}

	if e.quantifier != nil {
		e.Join(e.quantifier)
	}

	e.WriteByte('-')

	// Right arrow for right direction
	if e.direction == EdgeRight {
		e.WriteByte('>')
	}

	return e.String(), e.GetArgs()
}

func (e *EdgePattern) pattern() {}

// QuantifierType represents different quantifier types.
type QuantifierType int

const (
	QuantifierFixed QuantifierType = iota
	QuantifierBounded
)

// QuantifierBuilder is a builder for path quantifiers.
type QuantifierBuilder struct {
	sql.Builder
	quantifierType QuantifierType
	bound          int
	lowerBound     int
	upperBound     int
}

// NewFixedQuantifier creates a fixed quantifier {n}.
func NewFixedQuantifier(bound int) *QuantifierBuilder {
	return &QuantifierBuilder{
		quantifierType: QuantifierFixed,
		bound:          bound,
	}
}

// NewBoundedQuantifier creates a bounded quantifier {min,max}.
func NewBoundedQuantifier(lowerBound, upperBound int) *QuantifierBuilder {
	return &QuantifierBuilder{
		quantifierType: QuantifierBounded,
		lowerBound:     lowerBound,
		upperBound:     upperBound,
	}
}

// NewOpenBoundedQuantifier creates an open bounded quantifier {,max} (lower bound defaults to 0).
func NewOpenBoundedQuantifier(upperBound int) *QuantifierBuilder {
	return &QuantifierBuilder{
		quantifierType: QuantifierBounded,
		lowerBound:     0,
		upperBound:     upperBound,
	}
}

// Query returns the quantifier representation.
func (q *QuantifierBuilder) Query() (string, []any) {
	q.WrapBraces(func(b *sql.Builder) {
		switch q.quantifierType {
		case QuantifierFixed:
			b.WriteString(strconv.Itoa(q.bound))
		case QuantifierBounded:
			if q.lowerBound > 0 {
				b.WriteString(strconv.Itoa(q.lowerBound))
			}
			b.Comma()
			if q.upperBound > 0 {
				b.WriteString(strconv.Itoa(q.upperBound))
			}
		}
	})

	return q.String(), q.GetArgs()
}

// QuantifiedPatternBuilder is a builder for quantified path patterns.
type QuantifiedPatternBuilder struct {
	sql.Builder
	ptn        sql.Querier
	quantifier *QuantifierBuilder
}

// Quantified creates a new quantified pattern builder.
func Quantified() *QuantifiedPatternBuilder {
	return &QuantifiedPatternBuilder{}
}

// Pattern sets the pattern to be quantified.
func (q *QuantifiedPatternBuilder) Pattern(pattern sql.Querier) *QuantifiedPatternBuilder {
	q.ptn = pattern
	return q
}

// Quantifier sets the quantifier for the pattern.
func (q *QuantifiedPatternBuilder) Quantifier(quantifier *QuantifierBuilder) *QuantifiedPatternBuilder {
	q.quantifier = quantifier
	return q
}

// Fixed sets a fixed quantifier {n}.
func (q *QuantifiedPatternBuilder) Fixed(bound int) *QuantifiedPatternBuilder {
	q.quantifier = NewFixedQuantifier(bound)
	return q
}

// Bounded sets a bounded quantifier {min,max}.
func (q *QuantifiedPatternBuilder) Bounded(lowerBound, upperBound int) *QuantifiedPatternBuilder {
	q.quantifier = NewBoundedQuantifier(lowerBound, upperBound)
	return q
}

// OpenBounded sets an open bounded quantifier {,max}.
func (q *QuantifiedPatternBuilder) OpenBounded(upperBound int) *QuantifiedPatternBuilder {
	q.quantifier = NewOpenBoundedQuantifier(upperBound)
	return q
}

// Query returns the quantified pattern representation.
func (q *QuantifiedPatternBuilder) Query() (string, []any) {
	if q.ptn != nil {
		q.Join(q.ptn)
	}

	if q.quantifier != nil {
		quantQuery, quantArgs := q.quantifier.Query()
		q.WriteString(quantQuery)
		q.Args(quantArgs...)
	}

	return q.String(), q.GetArgs()
}

func (q *QuantifiedPatternBuilder) pattern() {}

// LabelExpr is a builder for complex label expressions.
type LabelExpr struct {
	sql.Builder
	depth int
	fns   []func(*sql.Builder)
}

// L creates a new label expression with the given label name.
func L(name string) *LabelExpr {
	return &LabelExpr{
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

func AndL(labels ...*LabelExpr) *LabelExpr {
	l := &LabelExpr{}
	return l.Append(func(b *sql.Builder) {
		l.mayWrap(labels, b, '&')
	})
}

func OrL(labels ...*LabelExpr) *LabelExpr {
	l := &LabelExpr{}
	return l.Append(func(b *sql.Builder) {
		l.mayWrap(labels, b, '|')
	})
}

func NotL(label *LabelExpr) *LabelExpr {
	l := &LabelExpr{}
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
func (l *LabelExpr) Append(fns ...func(*sql.Builder)) *LabelExpr {
	l.fns = append(l.fns, fns...)
	return l
}

func (l *LabelExpr) mayWrap(exprs []*LabelExpr, b *sql.Builder, op byte) {
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
func (l *LabelExpr) Query() (string, []any) {
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

// Predicate represents a GQL predicate expression.
type Predicate struct {
	sql.Builder
	depth int
	fns   []func(*sql.Builder)
}

// P creates a new GQL predicate.
func P(fns ...func(*sql.Builder)) *Predicate {
	return &Predicate{fns: fns}
}

// Raw allows injecting raw GQL expressions in the predicate.
func (p *Predicate) Raw(query string, args ...any) *Predicate {
	p.WriteString(query)
	p.Args(args...)
	return p
}

// And appends the AND operator to the predicate.
func (p *Predicate) And() *Predicate {
	p.WriteString(" AND ")
	return p
}

// Or appends the OR operator to the predicate.
func (p *Predicate) Or() *Predicate {
	p.WriteString(" OR ")
	return p
}

// Not wraps the predicate with the NOT operator.
func (p *Predicate) Not() *Predicate {
	p.Wrap(func(b *sql.Builder) {
		b.WriteString("NOT ")
		b.WriteString(p.String())
		b.Args(p.GetArgs()...)
	})
	return p
}

// Standard comparison predicates

// EQ adds a "=" predicate.
func (p *Predicate) EQ(column string, arg any) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpEQ).Arg(arg)
	})
}

// NEQ adds a "<>" predicate.
func (p *Predicate) NEQ(column string, arg any) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpNEQ).Arg(arg)
	})
}

// GT adds a ">" predicate.
func (p *Predicate) GT(column string, arg any) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpGT).Arg(arg)
	})
}

// GTE adds a ">=" predicate.
func (p *Predicate) GTE(column string, arg any) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpGTE).Arg(arg)
	})
}

// LT adds a "<" predicate.
func (p *Predicate) LT(column string, arg any) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpLT).Arg(arg)
	})
}

// LTE adds a "<=" predicate.
func (p *Predicate) LTE(column string, arg any) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpLTE).Arg(arg)
	})
}

// In adds an "IN" predicate.
func (p *Predicate) In(column string, args ...any) *Predicate {
	// If no arguments were provided, append the FALSE constant, since
	// we cannot apply "IN ()". This will make this predicate falsy.
	if len(args) == 0 {
		return p.False()
	}
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpIn)
		b.Wrap(func(b *sql.Builder) {
			if s, ok := args[0].(*sql.Selector); ok {
				b.Join(s)
			} else {
				b.Args(args...)
			}
		})
	})
}

// NotIn adds a "NOT IN" predicate.
func (p *Predicate) NotIn(column string, args ...any) *Predicate {
	// If no arguments were provided, append the NOT FALSE constant, since
	// we cannot apply "NOT IN ()". This will make this predicate truthy.
	if len(args) == 0 {
		return Not(p.False())
	}
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpNotIn)
		b.Wrap(func(b *sql.Builder) {
			if s, ok := args[0].(*sql.Selector); ok {
				b.Join(s)
			} else {
				b.Args(args...)
			}
		})
	})
}

// Like adds a "LIKE" predicate.
func (p *Predicate) Like(column, pattern string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpLike).Arg(pattern)
	})
}

// IsNull adds an "IS NULL" predicate.
func (p *Predicate) IsNull(column string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpIsNull)
	})
}

// NotNull adds an "IS NOT NULL" predicate.
func (p *Predicate) NotNull(column string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.Ident(column).WriteOp(sql.OpNotNull)
	})
}

// Graph-specific predicates

// IsLabeled adds an "IS LABELED" predicate.
func (p *Predicate) IsLabeled(element string, labelExpr *LabelExpr) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString(element).WriteString(" IS LABELED ")
		if labelExpr != nil {
			b.Join(labelExpr)
		}
	})
}

// IsNotLabeled adds an "IS NOT LABELED" predicate.
func (p *Predicate) IsNotLabeled(element string, labelExpr *LabelExpr) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString(element).WriteString(" IS NOT LABELED ")
		if labelExpr != nil {
			b.Join(labelExpr)
		}
	})
}

// IsSource adds an "IS SOURCE" predicate.
func (p *Predicate) IsSource(node, edge string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString(node).WriteString(" IS SOURCE ")
		if edge != "" {
			b.WriteString("OF ").WriteString(edge)
		}
	})
}

// IsNotSource adds an "IS NOT SOURCE" predicate.
func (p *Predicate) IsNotSource(node, edge string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString(node).WriteString(" IS NOT SOURCE ")
		if edge != "" {
			b.WriteString("OF ").WriteString(edge)
		}
	})
}

// IsDestination adds an "IS DESTINATION" predicate.
func (p *Predicate) IsDestination(node, edge string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString(node).WriteString(" IS DESTINATION ")
		if edge != "" {
			b.WriteString("OF ").WriteString(edge)
		}
	})
}

// IsNotDestination adds an "IS NOT DESTINATION" predicate.
func (p *Predicate) IsNotDestination(node, edge string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString(node).WriteString(" IS NOT DESTINATION ")
		if edge != "" {
			b.WriteString("OF ").WriteString(edge)
		}
	})
}

// Graph functions

// AllDifferent adds an "ALL_DIFFERENT" predicate.
func (p *Predicate) AllDifferent(elements ...string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString("ALL_DIFFERENT(")
		for i, element := range elements {
			if i > 0 {
				b.Comma()
			}
			b.WriteString(element)
		}
		b.WriteByte(')')
	})
}

// Same adds a "SAME" predicate.
func (p *Predicate) Same(elements ...string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString("SAME(")
		for i, element := range elements {
			if i > 0 {
				b.Comma()
			}
			b.WriteString(element)
		}
		b.WriteByte(')')
	})
}

// PropertyExists adds a "PROPERTY_EXISTS" predicate.
func (p *Predicate) PropertyExists(element, property string) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString("PROPERTY_EXISTS(").WriteString(element).Comma().WriteString(property).WriteByte(')')
	})
}

// EQ returns a "=" predicate.
func EQ(column string, arg any) *Predicate {
	return P().EQ(column, arg)
}

// NEQ returns a "<>" predicate.
func NEQ(column string, arg any) *Predicate {
	return P().NEQ(column, arg)
}

// GT returns a ">" predicate.
func GT(column string, arg any) *Predicate {
	return P().GT(column, arg)
}

// GTE returns a ">=" predicate.
func GTE(column string, arg any) *Predicate {
	return P().GTE(column, arg)
}

// LT returns a "<" predicate.
func LT(column string, arg any) *Predicate {
	return P().LT(column, arg)
}

// LTE returns a "<=" predicate.
func LTE(column string, arg any) *Predicate {
	return P().LTE(column, arg)
}

// In returns an "IN" predicate.
func In(column string, args ...any) *Predicate {
	return P().In(column, args...)
}

// NotIn returns a "NOT IN" predicate.
func NotIn(column string, args ...any) *Predicate {
	return P().NotIn(column, args...)
}

// Like returns a "LIKE" predicate.
func Like(column, pattern string) *Predicate {
	return P().Like(column, pattern)
}

// IsNull returns an "IS NULL" predicate.
func IsNull(column string) *Predicate {
	return P().IsNull(column)
}

// NotNull returns an "IS NOT NULL" predicate.
func NotNull(column string) *Predicate {
	return P().NotNull(column)
}

// Logical operations and utilities

// And combines all given predicates with AND between them.
func And(preds ...*Predicate) *Predicate {
	p := P()
	return p.Append(func(b *sql.Builder) {
		p.mayWrap(preds, b, "AND")
	})
}

// Or combines all given predicates with OR between them.
func Or(preds ...*Predicate) *Predicate {
	p := P()
	return p.Append(func(b *sql.Builder) {
		p.mayWrap(preds, b, "OR")
	})
}

// Not wraps the given predicate with the NOT operator.
func Not(pred *Predicate) *Predicate {
	return P().Not().Append(func(b *sql.Builder) {
		b.Wrap(func(b *sql.Builder) {
			b.Join(pred)
		})
	})
}

// False appends the FALSE keyword to the predicate.
func False() *Predicate {
	return P().False()
}

// False appends FALSE to the predicate.
func (p *Predicate) False() *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString("FALSE")
	})
}

// True appends the TRUE keyword to the predicate.
func True() *Predicate {
	return P().True()
}

// True appends TRUE to the predicate.
func (p *Predicate) True() *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString("TRUE")
	})
}

// ExprP creates a new predicate from the given expression.
func ExprP(expr string, args ...any) *Predicate {
	return P(func(b *sql.Builder) {
		b.Join(Expr(expr, args...))
	})
}

// Exists returns the EXISTS predicate.
func Exists(query sql.Querier) *Predicate {
	return P().Exists(query)
}

// Exists appends the EXISTS predicate with the given query.
func (p *Predicate) Exists(query sql.Querier) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString("EXISTS ")
		b.WrapBraces(func(b *sql.Builder) {
			b.Join(query)
		})
	})
}

// InSubquery adds an "IN" predicate with a subquery.
func (p *Predicate) InSubquery(value string, query sql.Querier) *Predicate {
	p.Ident(value).WriteOp(sql.OpIn)
	p.WrapBraces(func(b *sql.Builder) {
		b.Join(query)
	})
	return p
}

// NotInSubquery adds a "NOT IN" predicate with a subquery.
func (p *Predicate) NotInSubquery(value string, query sql.Querier) *Predicate {
	p.Ident(value).WriteOp(sql.OpNotIn)
	p.WrapBraces(func(b *sql.Builder) {
		b.Join(query)
	})
	return p
}

// NotExists returns the NOT EXISTS predicate.
func NotExists(query sql.Querier) *Predicate {
	return P().NotExists(query)
}

// InSubquery returns an "IN" predicate with a subquery.
func InSubquery(value string, query sql.Querier) *Predicate {
	return P().InSubquery(value, query)
}

// NotInSubquery returns a "NOT IN" predicate with a subquery.
func NotInSubquery(value string, query sql.Querier) *Predicate {
	return P().NotInSubquery(value, query)
}

// NotExists appends the NOT EXISTS predicate with the given query.
func (p *Predicate) NotExists(query sql.Querier) *Predicate {
	return p.Append(func(b *sql.Builder) {
		b.WriteString("NOT EXISTS ")
		b.Wrap(func(b *sql.Builder) {
			b.Join(query)
		})
	})
}

// Append appends a new function to the predicate callbacks.
// The callback list are executed on call to Query.
func (p *Predicate) Append(f func(*sql.Builder)) *Predicate {
	p.fns = append(p.fns, f)
	return p
}

// clone returns a shallow clone of p.
func (p *Predicate) clone() *Predicate {
	if p == nil {
		return p
	}
	return &Predicate{fns: append([]func(*sql.Builder){}, p.fns...)}
}

func (p *Predicate) mayWrap(preds []*Predicate, b *sql.Builder, op string) {
	switch n := len(preds); {
	case n == 1:
		b.Join(preds[0])
		return
	case n > 1 && p.depth != 0:
		b.WriteByte('(')
		defer b.WriteByte(')')
	}
	for i := range preds {
		preds[i].depth = p.depth + 1
		if i > 0 {
			b.WriteByte(' ')
			b.WriteString(op)
			b.WriteByte(' ')
		}
		if len(preds[i].fns) > 1 {
			b.Wrap(func(b *sql.Builder) {
				b.Join(preds[i])
			})
		} else {
			b.Join(preds[i])
		}
	}
}

// Query returns query representation of a predicate.
func (p *Predicate) Query() (string, []any) {
	if p.Len() > 0 || len(p.GetArgs()) > 0 {
		p.Reset()
		p.ClearArgs()
	}
	for _, f := range p.fns {
		f(&p.Builder)
	}
	return p.String(), p.GetArgs()
}

// arg calls Builder.Arg, but wraps complex queries with parens when needed.
func (*Predicate) arg(b *sql.Builder, a any) {
	switch a.(type) {
	case *GraphQuery, *Matcher, *FilterBuilder, *ReturnBuilder:
		// Wrap complex query types in parentheses
		b.Wrap(func(b *sql.Builder) {
			b.Arg(a)
		})
	default:
		b.Arg(a)
	}
}

// Graph-specific standalone predicates

// IsLabeled returns an "IS LABELED" predicate.
func IsLabeled(element string, labelExpr *LabelExpr) *Predicate {
	return P().IsLabeled(element, labelExpr)
}

// IsNotLabeled returns an "IS NOT LABELED" predicate.
func IsNotLabeled(element string, labelExpr *LabelExpr) *Predicate {
	return P().IsNotLabeled(element, labelExpr)
}

// IsSource returns an "IS SOURCE" predicate.
func IsSource(node, edge string) *Predicate {
	return P().IsSource(node, edge)
}

// IsNotSource returns an "IS NOT SOURCE" predicate.
func IsNotSource(node, edge string) *Predicate {
	return P().IsNotSource(node, edge)
}

// IsDestination returns an "IS DESTINATION" predicate.
func IsDestination(node, edge string) *Predicate {
	return P().IsDestination(node, edge)
}

// IsNotDestination returns an "IS NOT DESTINATION" predicate.
func IsNotDestination(node, edge string) *Predicate {
	return P().IsNotDestination(node, edge)
}

// AllDifferent returns an "ALL_DIFFERENT" predicate.
func AllDifferent(elements ...string) *Predicate {
	return P().AllDifferent(elements...)
}

// Same returns a "SAME" predicate.
func Same(elements ...string) *Predicate {
	return P().Same(elements...)
}

// PropertyExists returns a "PROPERTY_EXISTS" predicate.
func PropertyExists(element, property string) *Predicate {
	return P().PropertyExists(element, property)
}

// Raw returns a raw GQL query that is placed as-is in the query.
func Raw(s string) sql.Querier { return &raw{s} }

type raw struct{ s string }

func (r *raw) Query() (string, []any) { return r.s, nil }

// Expr returns an GQL expression that implements the sql.Querier interface.
func Expr(exr string, args ...any) sql.Querier { return &expr{s: exr, args: args} }

type expr struct {
	s    string
	args []any
}

func (e *expr) Query() (string, []any) { return e.s, e.args }

// ExprFunc returns an expression function that implements the sql.Querier interface.
func ExprFunc(fn func(*sql.Builder)) sql.Querier {
	return &exprFunc{fn: fn}
}

type exprFunc struct {
	sql.Builder
	fn func(*sql.Builder)
}

func (e *exprFunc) Query() (string, []any) {
	b := e.Builder.Clone()
	e.fn(&b)
	return b.Query()
}

// FuncBuilder is a builder for GQL functions.
type FuncBuilder struct {
	sql.Builder
	name string
	args []any
}

// Func creates a new function builder.
func Func(name string) *FuncBuilder {
	return &FuncBuilder{name: name}
}

// Args sets the arguments for the function.
func (f *FuncBuilder) Args(args ...any) *FuncBuilder {
	f.args = append(f.args, args...)
	return f
}

// Query returns the function call representation.
func (f *FuncBuilder) Query() (string, []any) {
	f.WriteString(f.name)
	f.WriteByte('(')
	f.Args(f.args...)
	f.WriteByte(')')
	return f.String(), f.GetArgs()
}

// DestinationNodeID returns a new DESTINATION_NODE_ID function builder.
func DestinationNodeID(edge any) *FuncBuilder {
	return Func("DESTINATION_NODE_ID").Args(edge)
}

// Edges returns a new EDGES function builder.
func Edges(path any) *FuncBuilder {
	return Func("EDGES").Args(path)
}

// ElementID returns a new ELEMENT_ID function builder.
func ElementID(element any) *FuncBuilder {
	return Func("ELEMENT_ID").Args(element)
}

// IsAcyclic returns a new IS_ACYCLIC function builder.
func IsAcyclic(path any) *FuncBuilder {
	return Func("IS_ACYCLIC").Args(path)
}

// IsSimple returns a new IS_SIMPLE function builder.
func IsSimple(path any) *FuncBuilder {
	return Func("IS_SIMPLE").Args(path)
}

// IsTrail returns a new IS_TRAIL function builder.
func IsTrail(path any) *FuncBuilder {
	return Func("IS_TRAIL").Args(path)
}

// LabelsFunc returns a new LABELS function builder.
func LabelsFunc(element any) *FuncBuilder {
	return Func("LABELS").Args(element)
}

// Nodes returns a new NODES function builder.
func Nodes(path any) *FuncBuilder {
	return Func("NODES").Args(path)
}

// PathFunc returns a new PATH function builder.
func PathFunc(elements ...any) *FuncBuilder {
	return Func("PATH").Args(elements...)
}

// PathFirst returns a new PATH_FIRST function builder.
func PathFirst(path any) *FuncBuilder {
	return Func("PATH_FIRST").Args(path)
}

// PathLast returns a new PATH_LAST function builder.
func PathLast(path any) *FuncBuilder {
	return Func("PATH_LAST").Args(path)
}

// PathLength returns a new PATH_LENGTH function builder.
func PathLength(path any) *FuncBuilder {
	return Func("PATH_LENGTH").Args(path)
}

// PropertyNames returns a new PROPERTY_NAMES function builder.
func PropertyNames(element any) *FuncBuilder {
	return Func("PROPERTY_NAMES").Args(element)
}

// SourceNodeID returns a new SOURCE_NODE_ID function builder.
func SourceNodeID(edge any) *FuncBuilder {
	return Func("SOURCE_NODE_ID").Args(edge)
}

// ArraySubqueryBuilder is a builder for ARRAY subqueries.
type ArraySubqueryBuilder struct {
	sql.Builder
	query sql.Querier
}

// ArraySubquery creates a new ARRAY subquery builder.
func ArraySubquery(query sql.Querier) *ArraySubqueryBuilder {
	return &ArraySubqueryBuilder{query: query}
}

// Query returns the ARRAY subquery representation.
func (a *ArraySubqueryBuilder) Query() (string, []any) {
	a.WriteString("ARRAY ")
	a.WrapBraces(func(b *sql.Builder) {
		b.Join(a.query)
	})
	return a.String(), a.GetArgs()
}

// ValueSubqueryBuilder is a builder for VALUE subqueries.
type ValueSubqueryBuilder struct {
	sql.Builder
	query sql.Querier
}

// ValueSubquery creates a new VALUE subquery builder.
func ValueSubquery(query sql.Querier) *ValueSubqueryBuilder {
	return &ValueSubqueryBuilder{query: query}
}

// Query returns the VALUE subquery representation.
func (v *ValueSubqueryBuilder) Query() (string, []any) {
	v.WriteString("VALUE ")
	v.WrapBraces(func(b *sql.Builder) {
		b.Join(v.query)
	})
	return v.String(), v.GetArgs()
}
