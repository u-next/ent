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
	schema     string              // property graph schema
	name       string              // property graph name
	exists     bool                // check existence
	nodeTablex []*NodeTableBuilder // node table builders
	edgeTablex []*EdgeTableBuilder // edge table builders
	options    []string            // property graph options
	comment    string              // property graph comment
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

// Schema sets the database schema name for the property graph.
func (pg *PropertyGraphBuilder) Schema(name string) *PropertyGraphBuilder {
	pg.schema = name
	return pg
}

// IfNotExists appends the `IF NOT EXISTS` clause to the `CREATE PROPERTY GRAPH` statement.
func (pg *PropertyGraphBuilder) IfNotExists() *PropertyGraphBuilder {
	pg.exists = true
	return pg
}

// NodeTable appends node tables to the property graph.
func (pg *PropertyGraphBuilder) NodeTable(nts ...*NodeTableBuilder) *PropertyGraphBuilder {
	pg.nodeTablex = append(pg.nodeTablex, nts...)
	return pg
}

// EdgeTable appends edge tables to the property graph.
func (pg *PropertyGraphBuilder) EdgeTable(ets ...*EdgeTableBuilder) *PropertyGraphBuilder {
	pg.edgeTablex = append(pg.edgeTablex, ets...)
	return pg
}

// Options sets the property graph options.
func (pg *PropertyGraphBuilder) Options(opts ...string) *PropertyGraphBuilder {
	pg.options = append(pg.options, opts...)
	return pg
}

// Comment sets the property graph comment.
func (pg *PropertyGraphBuilder) Comment(comment string) *PropertyGraphBuilder {
	pg.comment = comment
	return pg
}

// Query returns query representation of a `CREATE PROPERTY GRAPH` statement.
func (pg *PropertyGraphBuilder) Query() (string, []any) {
	pg.WriteString("CREATE PROPERTY GRAPH ")
	if pg.exists {
		pg.WriteString("IF NOT EXISTS ")
	}
	pg.writeSchema(pg.schema)
	pg.Ident(pg.name)

	if len(pg.nodeTablex) > 0 || len(pg.edgeTablex) > 0 {
		pg.NewLine().WriteString("NODE TABLES (")
		for i, nt := range pg.nodeTablex {
			if i > 0 {
				pg.Comma()
			}
			pg.NewLine().Indent(1)
			pg.writeNodeTable(nt)
		}
		pg.NewLine().WriteByte(')')

		if len(pg.edgeTablex) > 0 {
			pg.NewLine().WriteString("EDGE TABLES (")
			for i, et := range pg.edgeTablex {
				if i > 0 {
					pg.Comma()
				}
				pg.NewLine().Indent(1)
				pg.writeEdgeTable(et)
			}
			pg.NewLine().WriteByte(')')
		}
	}

	if len(pg.options) > 0 {
		pg.NewLine().WriteString("OPTIONS (")
		for i, opt := range pg.options {
			if i > 0 {
				pg.Comma()
			}
			pg.WriteString(opt)
		}
		pg.WriteByte(')')
	}

	return pg.String(), pg.GetArgs()
}

// writeSchema writes the schema prefix if provided.
func (pg *PropertyGraphBuilder) writeSchema(schema string) {
	if schema != "" {
		pg.Ident(schema).WriteByte('.')
	}
}

// writeNodeTable writes a node table definition.
func (pg *PropertyGraphBuilder) writeNodeTable(nt *NodeTableBuilder) {
	pg.Ident(nt.tableName)
	if nt.alias != "" {
		pg.WriteString(" AS ").Ident(nt.alias)
	}

	if nt.key != nil {
		pg.NewLine().Indent(2).WriteString("KEY (")
		for i, col := range nt.key.Columns {
			if i > 0 {
				pg.Comma()
			}
			pg.Ident(col)
		}
		pg.WriteByte(')')
	}

	if len(nt.labels) > 0 {
		pg.NewLine().Indent(2).WriteString("LABEL (")
		for i, label := range nt.labels {
			if i > 0 {
				pg.Comma()
			}
			pg.WriteString(label.Name)
			if len(label.Properties) > 0 {
				pg.WriteString(" PROPERTIES (")
				for j, prop := range label.Properties {
					if j > 0 {
						pg.Comma()
					}
					pg.Ident(prop)
				}
				pg.WriteByte(')')
			}
		}
		pg.WriteByte(')')
	}
}

// writeEdgeTable writes an edge table definition.
func (pg *PropertyGraphBuilder) writeEdgeTable(et *EdgeTableBuilder) {
	pg.Ident(et.tableName)
	if et.alias != "" {
		pg.WriteString(" AS ").Ident(et.alias)
	}

	if et.key != nil {
		pg.NewLine().Indent(2).WriteString("KEY (")
		for i, col := range et.key.Columns {
			if i > 0 {
				pg.Comma()
			}
			pg.Ident(col)
		}
		pg.WriteByte(')')
	}

	if et.sourceKey != nil {
		pg.NewLine().Indent(2).WriteString("SOURCE KEY (")
		for i, col := range et.sourceKey.Columns {
			if i > 0 {
				pg.Comma()
			}
			pg.Ident(col)
		}
		pg.WriteString(") REFERENCES ").Ident(et.sourceKey.ReferencedTable)
		if len(et.sourceKey.ReferencedColumns) > 0 {
			pg.WriteString(" (")
			for i, col := range et.sourceKey.ReferencedColumns {
				if i > 0 {
					pg.Comma()
				}
				pg.Ident(col)
			}
			pg.WriteByte(')')
		}
		pg.WriteByte(')')
	}

	if et.destinationKey != nil {
		pg.NewLine().Indent(2).WriteString("DESTINATION KEY (")
		for i, col := range et.destinationKey.Columns {
			if i > 0 {
				pg.Comma()
			}
			pg.Ident(col)
		}
		pg.WriteString(") REFERENCES ").Ident(et.destinationKey.ReferencedTable)
		if len(et.destinationKey.ReferencedColumns) > 0 {
			pg.WriteString(" (")
			for i, col := range et.destinationKey.ReferencedColumns {
				if i > 0 {
					pg.Comma()
				}
				pg.Ident(col)
			}
			pg.WriteByte(')')
		}
		pg.WriteByte(')')
	}

	if len(et.labels) > 0 {
		pg.NewLine().Indent(2).WriteString("LABEL (")
		for i, label := range et.labels {
			if i > 0 {
				pg.Comma()
			}
			pg.WriteString(label.Name)
			if len(label.Properties) > 0 {
				pg.WriteString(" PROPERTIES (")
				for j, prop := range label.Properties {
					if j > 0 {
						pg.Comma()
					}
					pg.Ident(prop)
				}
				pg.WriteByte(')')
			}
		}
		pg.WriteByte(')')
	}
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
