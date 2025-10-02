// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Package sqlpgq provides property graph query building support for SQL dialects.
package sqlpgq

import (
	"entgo.io/ent/dialect/sql"
)

// ElementKey represents an element key definition.
type ElementKey struct {
	Columns []string // Column names that form the key
}

// NewElementKey returns a new element key with the given columns.
func NewElementKey(columns ...string) *ElementKey {
	return &ElementKey{
		Columns: columns,
	}
}

// ReferenceKey represents a source or destination key definition for edges.
type ReferenceKey struct {
	Columns           []string // Column names in the edge table
	ReferencedTable   string   // Name of the referenced table (node table)
	ReferencedColumns []string // Column names in the referenced table
}

// NewReferenceKey returns a new reference key.
func NewReferenceKey(columns []string, refTable string, refColumns []string) *ReferenceKey {
	return &ReferenceKey{
		Columns:           columns,
		ReferencedTable:   refTable,
		ReferencedColumns: refColumns,
	}
}

// Label represents a label definition for nodes or edges.
type Label struct {
	Name       string   // Label name
	Properties []string // Simple property list
}

// NewLabel returns a new label with the given name.
func NewLabel(name string) *Label {
	return &Label{
		Name: name,
	}
}

// PropertyGraphBuilder is a builder for `CREATE PROPERTY GRAPH` statement.
type PropertyGraphBuilder struct {
	sql.Builder
	schema     string              // optional schema prefix
	name       string              // property graph name
	exists     bool                // check existence
	replace    bool                // replace existing property graph
	nodeTables []*NodeTableBuilder // node table builders
	edgeTables []*EdgeTableBuilder // edge table builders
}

// CreatePropertyGraph returns a query builder for the `CREATE PROPERTY GRAPH` statement.
//
//	CreatePropertyGraph("my_graph").
//		NodeTable(
//			NodeTable("users").Key("id").Label("Person"),
//			NodeTable("posts").Key("id").Label("Post"),
//		).
//		EdgeTable(
//			EdgeTable("follows").Key("id").SourceKey(...),
//			EdgeTableFromSQL(sql.Table("likes")),
//		).
//		Options("HINT_1=value1", "HINT_2=value2")
func CreatePropertyGraph(name string) *PropertyGraphBuilder {
	return &PropertyGraphBuilder{name: name}
}

// IfNotExists appends the `IF NOT EXISTS` clause to the `CREATE PROPERTY GRAPH` statement.
func (pg *PropertyGraphBuilder) IfNotExists() *PropertyGraphBuilder {
	pg.exists = true
	return pg
}

// OrReplace appends the `OR REPLACE` clause to the `CREATE PROPERTY GRAPH` statement.
func (pg *PropertyGraphBuilder) OrReplace() *PropertyGraphBuilder {
	pg.replace = true
	return pg
}

// NodeTable appends node tables to the property graph.
func (pg *PropertyGraphBuilder) NodeTable(nts ...*NodeTableBuilder) *PropertyGraphBuilder {
	pg.nodeTables = append(pg.nodeTables, nts...)
	return pg
}

// EdgeTable appends edge tables to the property graph.
func (pg *PropertyGraphBuilder) EdgeTable(ets ...*EdgeTableBuilder) *PropertyGraphBuilder {
	pg.edgeTables = append(pg.edgeTables, ets...)
	return pg
}

// SetSchema sets the optional schema prefix for the property graph.
func (pg *PropertyGraphBuilder) SetSchema(schema string) *PropertyGraphBuilder {
	pg.schema = schema
	return pg
}

func (pg *PropertyGraphBuilder) Name() string {
	return pg.name
}

// Query returns query representation of a `CREATE PROPERTY GRAPH` statement.
func (pg *PropertyGraphBuilder) Query() (string, []any) {
	pg.WriteString("CREATE ")
	if pg.replace {
		pg.WriteString("OR REPLACE ")
	}
	pg.WriteString("PROPERTY GRAPH ")
	if pg.exists {
		pg.WriteString("IF NOT EXISTS ")
	}
	pg.WriteSchema(pg.schema)
	pg.Ident(pg.name)

	// Add node and edge tables
	if len(pg.nodeTables) > 0 || len(pg.edgeTables) > 0 {
		pg.Indent()
		pg.NewLine().WriteString("NODE TABLES ")
		pg.Wrap(func(b *sql.Builder) {
			pg.Indent()
			for i, nt := range pg.nodeTables {
				if i > 0 {
					pg.Comma()
				}
				pg.NewLine()
				pg.Join(nt)
			}
			pg.Dedent().NewLine()
		})
		if len(pg.edgeTables) > 0 {
			pg.NewLine().WriteString("EDGE TABLES ")
			pg.Wrap(func(b *sql.Builder) {
				pg.Indent()
				for i, et := range pg.edgeTables {
					if i > 0 {
						pg.Comma()
					}
					pg.NewLine()
					pg.writeEdgeTable(et)
				}
				pg.Dedent().NewLine()
			})
		}
		pg.Dedent()
		pg.WriteByte(';')
	}

	return pg.String(), pg.GetArgs()
}

// writeEdgeTable writes an edge table definition.
func (pg *PropertyGraphBuilder) writeEdgeTable(et *EdgeTableBuilder) {

}

// NodeTable returns a new node table builder.
func NodeTable(tableName string) *NodeTableBuilder {
	return &NodeTableBuilder{tableName: tableName}
}

// NodeTableFromSQL creates a node table builder from a sql.SelectTable.
// The table name becomes the node table name, and the alias (if any) becomes a label.
// Example: users AS user_table -> NodeTable("users").Label("user_table")
func NodeTableFromSQL(table *sql.SelectTable) *NodeTableBuilder {
	nt := &NodeTableBuilder{tableName: table.Name()}

	// If the table has an alias, use it as a label
	if alias := table.Alias(); alias != "" {
		nt.alias = alias
		// Also add the alias as a default label
		label := NewLabel(alias)
		nt.labels = append(nt.labels, label)
	}

	return nt
}

// NodeTableBuilder builds node table definitions.
type NodeTableBuilder struct {
	sql.Builder
	tableName string
	alias     string
	key       *ElementKey
	labels    []*Label
}

// As sets an alias for the node table.
func (nt *NodeTableBuilder) As(alias string) *NodeTableBuilder {
	nt.alias = alias
	return nt
}

// Key sets the element key for the node table.
func (nt *NodeTableBuilder) Key(columns ...string) *NodeTableBuilder {
	nt.key = NewElementKey(columns...)
	return nt
}

// Label adds a label to the node table.
func (nt *NodeTableBuilder) Label(name string) *LabelBuilder {
	label := NewLabel(name)
	nt.labels = append(nt.labels, label)
	return &LabelBuilder{label: label}
}

func (nt *NodeTableBuilder) Query() (string, []any) {
	nt.Ident(nt.tableName)
	if nt.alias != "" {
		nt.WriteString(" AS ").Ident(nt.alias)
	}

	nt.Indent()
	if nt.key != nil {
		nt.NewLine().WriteString("KEY ")
		nt.Wrap(func(b *sql.Builder) {
			nt.IdentComma(nt.key.Columns...)
		})
	}

	for _, label := range nt.labels {
		nt.NewLine().WriteString("LABEL ")
		nt.WriteString(label.Name)
		if len(label.Properties) > 0 {
			nt.WriteString(" PROPERTIES ")
			nt.Wrap(func(b *sql.Builder) {
				nt.IdentComma(label.Properties...)
			})
		}
	}
	nt.Dedent()

	return nt.String(), nt.GetArgs()
}

// ToSelectTable converts the node table to a sql.SelectTable.
// If the node table has an alias that differs from the table name, it's applied to the SelectTable.
// If there's a label with the same name as the table, no alias is set.
func (nt *NodeTableBuilder) ToSelectTable() *sql.SelectTable {
	table := sql.Table(nt.tableName)

	// Only set alias if it's different from the table name
	if nt.alias != "" && nt.alias != nt.tableName {
		// Check if there's a label with the same name as the table name
		hasTableNameLabel := false
		for _, label := range nt.labels {
			if label.Name == nt.tableName {
				hasTableNameLabel = true
				break
			}
		}

		// If there's no label matching the table name, use the alias
		if !hasTableNameLabel {
			table = table.As(nt.alias)
		}
	}

	return table
}

func (nt *NodeTableBuilder) L() *LabelExpr {
	if nt.alias != "" {
		return L(nt.alias)
	}
	return L(nt.tableName)
}

// EdgeTable returns a new edge table builder.
func EdgeTable(tableName string) *EdgeTableBuilder {
	return &EdgeTableBuilder{tableName: tableName}
}

// EdgeTableFromSQL creates an edge table builder from a sql.SelectTable.
// The table name becomes the edge table name, and the alias (if any) becomes a label.
// Example: follows AS follow_edge -> EdgeTable("follows").Label("follow_edge")
func EdgeTableFromSQL(table *sql.SelectTable) *EdgeTableBuilder {
	et := &EdgeTableBuilder{tableName: table.Name()}

	// If the table has an alias, use it as a label
	if alias := table.Alias(); alias != "" {
		et.alias = alias
		// Also add the alias as a default label
		label := NewLabel(alias)
		et.labels = append(et.labels, label)
	}

	return et
}

// EdgeTableBuilder builds edge table definitions.
type EdgeTableBuilder struct {
	sql.Builder
	tableName      string
	alias          string
	key            *ElementKey
	sourceKey      *ReferenceKey
	destinationKey *ReferenceKey
	labels         []*Label
}

// As sets an alias for the edge table.
func (et *EdgeTableBuilder) As(alias string) *EdgeTableBuilder {
	et.alias = alias
	return et
}

// Key sets the element key for the edge table.
func (et *EdgeTableBuilder) Key(columns ...string) *EdgeTableBuilder {
	et.key = NewElementKey(columns...)
	return et
}

// SourceKey sets the source key for the edge table.
func (et *EdgeTableBuilder) SourceKey(columns []string, referencedTable string, referencedColumns ...string) *EdgeTableBuilder {
	et.sourceKey = NewReferenceKey(columns, referencedTable, referencedColumns)
	return et
}

// DestinationKey sets the destination key for the edge table.
func (et *EdgeTableBuilder) DestinationKey(columns []string, referencedTable string, referencedColumns ...string) *EdgeTableBuilder {
	et.destinationKey = NewReferenceKey(columns, referencedTable, referencedColumns)
	return et
}

// Label adds a label to the edge table.
func (et *EdgeTableBuilder) Label(name string) *LabelBuilder {
	label := NewLabel(name)
	et.labels = append(et.labels, label)
	return &LabelBuilder{label: label}
}

// Query returns query representation of the edge table.
func (et *EdgeTableBuilder) Query() (string, []any) {
	et.Ident(et.tableName)
	if et.alias != "" {
		et.WriteString(" AS ").Ident(et.alias)
	}

	et.Indent()
	if et.key != nil {
		et.NewLine().WriteString("KEY ")
		et.Wrap(func(b *sql.Builder) {
			et.IdentComma(et.key.Columns...)
		})
	}
	if et.sourceKey != nil {
		et.NewLine().WriteString("SOURCE KEY ")
		et.Wrap(func(b *sql.Builder) {
			et.IdentComma(et.sourceKey.Columns...)
		})
		if et.sourceKey.ReferencedTable != "" && len(et.sourceKey.ReferencedColumns) > 0 {
			et.WriteString(" REFERENCES ")
			et.Ident(et.sourceKey.ReferencedTable)
			et.WriteByte(' ')
			et.Wrap(func(b *sql.Builder) {
				et.IdentComma(et.sourceKey.ReferencedColumns...)
			})
		}
	}
	if et.destinationKey != nil {
		et.NewLine().WriteString("DESTINATION KEY ")
		et.Wrap(func(b *sql.Builder) {
			et.IdentComma(et.destinationKey.Columns...)
		})
		if et.destinationKey.ReferencedTable != "" && len(et.destinationKey.ReferencedColumns) > 0 {
			et.WriteString(" REFERENCES ")
			et.Ident(et.destinationKey.ReferencedTable)
			et.WriteByte(' ')
			et.Wrap(func(b *sql.Builder) {
				et.IdentComma(et.destinationKey.ReferencedColumns...)
			})
		}
	}

	for _, label := range et.labels {
		et.NewLine().WriteString("LABEL ")
		et.WriteString(label.Name)
		if len(label.Properties) > 0 {
			et.WriteString(" PROPERTIES ")
			et.Wrap(func(b *sql.Builder) {
				et.IdentComma(label.Properties...)
			})
		}
	}
	et.Dedent()

	return et.String(), et.GetArgs()
}

// ToSelectTable converts the edge table to a sql.SelectTable.
// If the edge table has an alias that differs from the table name, it's applied to the SelectTable.
// If there's a label with the same name as the table, no alias is set.
func (et *EdgeTableBuilder) ToSelectTable() *sql.SelectTable {
	table := sql.Table(et.tableName)

	// Only set alias if it's different from the table name
	if et.alias != "" && et.alias != et.tableName {
		// Check if there's a label with the same name as the table name
		hasTableNameLabel := false
		for _, label := range et.labels {
			if label.Name == et.tableName {
				hasTableNameLabel = true
				break
			}
		}

		// If there's no label matching the table name, use the alias
		if !hasTableNameLabel {
			table = table.As(et.alias)
		}
	}

	return table
}

func (et *EdgeTableBuilder) L() *LabelExpr {
	if et.alias != "" {
		return L(et.alias)
	}
	return L(et.tableName)
}

// LabelBuilder builds label definitions.
type LabelBuilder struct {
	label *Label
}

// Properties sets the properties for the label.
func (lb *LabelBuilder) Properties(properties ...string) *LabelBuilder {
	lb.label.Properties = append(lb.label.Properties, properties...)
	return lb
}
