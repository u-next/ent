package sqlpgq

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlhint"
)

// GraphQuery is a builder for complete GQL queries.
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

// FIXME: refactor order by clauses
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

// FIXME: Const creates a new constant column expression.
func Const(v any) *columnExpr {
	return &columnExpr{expr: sql.Expr("?", v)}
}

// Column creates a new column expression.
func Column(expr string, args ...any) *columnExpr {
	return &columnExpr{expr: sql.Expr(expr, args...)}
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
	if len(r.items) > 0 {
		r.Pad()
		r.JoinComma(r.items...)
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
	sql.Builder
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
		b.WriteByte('.')
		b.Ident(f)
	}
	return b.String()
}

func (a *assignment) Query() (string, []any) {
	a.Ident(a.variable)
	a.WriteString(" = ")
	a.Join(a.value)
	return a.String(), a.GetArgs()
}

// LetBuilder is a builder for LET statements.
type LetBuilder struct {
	sql.Builder
	assignments []sql.Querier
}

// Let creates a new LET statement builder.
func Let() *LetBuilder {
	return &LetBuilder{}
}

// Append adds more assignments to the LET statement.
func (l *LetBuilder) Append(as ...*assignment) *LetBuilder {
	for _, a := range as {
		l.assignments = append(l.assignments, a)
	}
	return l
}

// Query returns the LET statement representation.
func (l *LetBuilder) Query() (string, []any) {
	l.WriteString("LET")
	if len(l.assignments) > 0 {
		l.Pad()
		l.JoinComma(l.assignments...)
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
	g.WriteString("GROUP BY")
	if len(g.exprs) > 0 {
		g.Pad()
		g.JoinComma(g.exprs...)
	}
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
	o.WriteString("ORDER BY")
	if len(o.orders) > 0 {
		o.Pad()
		o.JoinComma(o.orders...)
	}
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
	f.WriteString("FOR")
	if f.element != "" {
		f.Pad()
		f.Ident(f.element)
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
	if len(w.items) > 0 {
		w.Pad()
		w.JoinComma(w.items...)
	}
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
		b.WriteString("UNION ALL")
	}
}

func UnionDistinct() SetOperation {
	return func(b *sql.Builder) {
		b.WriteString("UNION DISTINCT")
	}
}

func IntersectAll() SetOperation {
	return func(b *sql.Builder) {
		b.WriteString("INTERSECT ALL")
	}
}

func IntersectDistinct() SetOperation {
	return func(b *sql.Builder) {
		b.WriteString("INTERSECT DISTINCT")
	}
}

func ExceptAll() SetOperation {
	return func(b *sql.Builder) {
		b.WriteString("EXCEPT ALL")
	}
}

func ExceptDistinct() SetOperation {
	return func(b *sql.Builder) {
		b.WriteString("EXCEPT DISTINCT")
	}
}

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

// Concat concatenates two path patterns with the || operator.
func Concat(p, q *PathPattern) sql.Querier {
	return sql.ExprFunc(func(b *sql.Builder) {
		if p.variable != "" {
			b.Ident(p.variable)
		} else {
			b.Wrap(func(b *sql.Builder) {
				b.Join(p)
			})
		}
		b.WriteString(" || ")
		if q.variable != "" {
			b.Ident(q.variable)
		} else {
			b.Wrap(func(b *sql.Builder) {
				b.Join(q)
			})
		}
	})
}

type GraphPatternBuilder struct {
	sql.Builder
	patterns []sql.Querier
	where    *sql.Predicate
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

func (w *GraphPatternBuilder) Query() (string, []any) {
	w.JoinComma(w.patterns...)
	w.NewLine()
	if w.where != nil {
		w.WriteString("WHERE ")
		w.Join(w.where)
	}
	return w.String(), w.GetArgs()
}

func (w *GraphPatternBuilder) pattern() {}

// Pattern is an interface for graph patterns.
type Pattern interface {
	sql.Querier
	pattern()
}

// PathPattern is a builder for path patterns.
type PathPattern struct {
	sql.Builder
	variable     string
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
	b := p.Clone()
	quantified := p.bound[0] != nil || p.bound[1] != nil
	if quantified {
		b.WriteByte('(')
	}
	if p.variable != "" {
		b.Ident(p.variable)
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
	filler *patternFiller
}

// N creates a new node pattern builder.
func N() *NodePattern {
	return &NodePattern{
		filler: &patternFiller{properties: make(map[string]sql.Querier)},
	}
}

// Named sets the node variable.
func (n *NodePattern) Named(name string) *NodePattern {
	n.filler.variable = name
	return n
}

// Labels sets the node labels using simple OR logic.
func (n *NodePattern) Labels(labels ...string) *NodePattern {
	ls := make([]*LabelExpr, len(labels))
	for i, l := range labels {
		ls[i] = L(l)
	}
	n.filler.labelExpr = OrL(ls...)
	return n
}

// LabelExpression sets a complex label expression.
func (n *NodePattern) LabelExpr(expr *LabelExpr) *NodePattern {
	n.filler.labelExpr = expr
	return n
}

// Property adds a property filter.
func (n *NodePattern) Property(key string, expr sql.Querier) *NodePattern {
	n.filler.properties[key] = expr
	return n
}

// Properties adds multiple property filters.
func (n *NodePattern) Properties(props map[string]sql.Querier) *NodePattern {
	maps.Copy(n.filler.properties, props)
	return n
}

// Where adds a WHERE condition.
func (n *NodePattern) Where(condition *sql.Predicate) *NodePattern {
	n.filler.where = condition
	return n
}

// F returns a formatted string for the field of the node variable.
func (n *NodePattern) F(fields ...string) string {
	var b sql.Builder
	if n.filler.variable != "" {
		b.Ident(n.filler.variable)
	}
	for _, field := range fields {
		b.WriteByte('.').Ident(field)
	}
	return b.String()
}

// L returns a label expression for the node pattern.
func (n *NodePattern) L() *LabelExpr {
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
}

// E creates a new edge pattern builder.
func E() *EdgePattern {
	return &EdgePattern{
		direction: EdgeAnyDirection,
		filler:    &patternFiller{properties: make(map[string]sql.Querier)},
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
	e.filler.variable = name
	return e
}

// Labels sets the edge labels using simple OR logic.
func (e *EdgePattern) Labels(labels ...string) *EdgePattern {
	ls := make([]*LabelExpr, len(labels))
	for i, l := range labels {
		ls[i] = L(l)
	}
	e.filler.labelExpr = OrL(ls...)
	return e
}

// LabelExpr sets a complex label expression.
func (e *EdgePattern) LabelExpr(expr *LabelExpr) *EdgePattern {
	e.filler.labelExpr = expr
	return e
}

// Property adds a property filter.
func (e *EdgePattern) Property(key string, value sql.Querier) *EdgePattern {
	e.filler.properties[key] = value
	return e
}

// Properties adds multiple property filters.
func (e *EdgePattern) Properties(props map[string]sql.Querier) *EdgePattern {
	maps.Copy(e.filler.properties, props)
	return e
}

// Where adds a WHERE condition.
func (e *EdgePattern) Where(condition sql.Querier) *EdgePattern {
	e.filler.where = condition
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
	if e.filler.variable != "" {
		b.Ident(e.filler.variable)
	}
	for _, field := range fields {
		b.WriteByte('.').Ident(field)
	}
	return b.String()
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

func (e *EdgePattern) pattern() {}

type patternFiller struct {
	sql.Builder
	variable   string
	labelExpr  *LabelExpr
	properties map[string]sql.Querier
	where      sql.Querier
}

func (p *patternFiller) Query() (string, []any) {
	b := p.Clone()
	if p.variable != "" {
		b.Ident(p.variable)
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
func IsLabeled(element string, labelExpr *LabelExpr) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.Ident(element)
		b.WriteString(" IS LABELED ")
		if labelExpr != nil {
			b.Join(labelExpr)
		}
	})
}

// IsNotLabeled returns an "IS NOT LABELED" predicate.
func IsNotLabeled(element string, labelExpr *LabelExpr) *sql.Predicate {
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
			for i, element := range elements {
				if i > 0 {
					b.Comma()
				}
				b.WriteString(element)
			}
		})
	})
}

// Same returns a "SAME" predicate.
func Same(elements ...string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("SAME")
		b.Wrap(func(b *sql.Builder) {
			for i, element := range elements {
				if i > 0 {
					b.Comma()
				}
				b.WriteString(element)
			}
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
