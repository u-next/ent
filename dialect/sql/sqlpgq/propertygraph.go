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
	Name       string            // Label name
	IsDefault  bool              // Whether this is the default label
	Properties *PropertiesConfig // Properties configuration
}

// NewLabel returns a new label with the given name.
func NewLabel(name string) *Label {
	return &Label{
		Name: name,
	}
}

// NewDefaultLabel returns a new default label.
func NewDefaultLabel() *Label {
	return &Label{
		IsDefault: true,
	}
}

// SetProperties sets the properties configuration for this label.
func (l *Label) SetProperties(props *PropertiesConfig) *Label {
	l.Properties = props
	return l
}

// PropertiesConfig represents different types of property configurations.
type PropertiesConfig struct {
	Type              PropertiesType     // Type of properties definition
	ExceptColumns     []string           // Columns to exclude when using ALL COLUMNS
	DerivedProperties []*DerivedProperty // List of derived properties
}

// PropertiesType represents the type of properties definition.
type PropertiesType int

const (
	PropertiesTypeNone    PropertiesType = iota // NO PROPERTIES
	PropertiesTypeAll                           // PROPERTIES ARE ALL COLUMNS
	PropertiesTypeDerived                       // PROPERTIES (derived_property_list)
)

// NewNoProperties returns a properties configuration with no properties.
func NewNoProperties() *PropertiesConfig {
	return &PropertiesConfig{
		Type: PropertiesTypeNone,
	}
}

// NewAllProperties returns a properties configuration that includes all columns.
func NewAllProperties() *PropertiesConfig {
	return &PropertiesConfig{
		Type: PropertiesTypeAll,
	}
}

// NewAllPropertiesExcept returns a properties configuration that includes all columns except the specified ones.
func NewAllPropertiesExcept(exceptColumns ...string) *PropertiesConfig {
	return &PropertiesConfig{
		Type:          PropertiesTypeAll,
		ExceptColumns: exceptColumns,
	}
}

// NewDerivedProperties returns a properties configuration with derived properties.
func NewDerivedProperties(props ...*DerivedProperty) *PropertiesConfig {
	return &PropertiesConfig{
		Type:              PropertiesTypeDerived,
		DerivedProperties: props,
	}
}

// DerivedProperty represents a derived property definition.
type DerivedProperty struct {
	Expression string // Value expression for the property
	Alias      string // Optional alias for the property
}

// NewDerivedProperty returns a new derived property.
func NewDerivedProperty(expression string) *DerivedProperty {
	return &DerivedProperty{
		Expression: expression,
	}
}

// NewDerivedPropertyAs returns a new derived property with an alias.
func NewDerivedPropertyAs(expression, alias string) *DerivedProperty {
	return &DerivedProperty{
		Expression: expression,
		Alias:      alias,
	}
}

// DynamicLabel represents a dynamic label definition.
type DynamicLabel struct {
	ColumnName string // Name of the column that holds label values
}

// NewDynamicLabel returns a new dynamic label definition.
func NewDynamicLabel(columnName string) *DynamicLabel {
	return &DynamicLabel{
		ColumnName: columnName,
	}
}

// DynamicProperties represents a dynamic properties definition.
type DynamicProperties struct {
	ColumnName string // Name of the column that holds properties values (must be JSON)
}

// NewDynamicProperties returns a new dynamic properties definition.
func NewDynamicProperties(columnName string) *DynamicProperties {
	return &DynamicProperties{
		ColumnName: columnName,
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
	b := pg.Clone()
	b.WriteString("CREATE ")
	if pg.replace {
		b.WriteString("OR REPLACE ")
	}
	b.WriteString("PROPERTY GRAPH ")
	if pg.exists {
		b.WriteString("IF NOT EXISTS ")
	}
	b.WriteSchema(pg.schema)
	b.Ident(pg.name)

	// Add node and edge tables
	if len(pg.nodeTables) > 0 || len(pg.edgeTables) > 0 {
		b.Indent()
		b.NewLine().WriteString("NODE TABLES ")
		b.Wrap(func(b *sql.Builder) {
			b.Indent()
			for i, nt := range pg.nodeTables {
				if i > 0 {
					b.Comma()
				}
				b.NewLine()
				b.Join(nt)
			}
			b.Dedent().NewLine()
		})
		if len(pg.edgeTables) > 0 {
			b.NewLine().WriteString("EDGE TABLES ")
			b.Wrap(func(b *sql.Builder) {
				b.Indent()
				for i, et := range pg.edgeTables {
					if i > 0 {
						b.Comma()
					}
					b.NewLine()
					b.Join(et)
				}
				b.Dedent().NewLine()
			})
		}
		b.Dedent()
	}

	return b.String(), b.GetArgs()
}

// writeEdgeTable writes an edge table definition.
func (pg *PropertyGraphBuilder) writeEdgeTable(et *EdgeTableBuilder) {
	pg.Join(et)
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
	tableName         string
	alias             string
	key               *ElementKey
	labels            []*Label
	dynamicLabel      *DynamicLabel
	dynamicProperties *DynamicProperties
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

// DefaultLabel adds a default label to the node table.
func (nt *NodeTableBuilder) DefaultLabel() *LabelBuilder {
	label := NewDefaultLabel()
	nt.labels = append(nt.labels, label)
	return &LabelBuilder{label: label}
}

// DynamicLabel sets the dynamic label for the node table.
func (nt *NodeTableBuilder) DynamicLabel(columnName string) *NodeTableBuilder {
	nt.dynamicLabel = NewDynamicLabel(columnName)
	return nt
}

// DynamicProperties sets the dynamic properties for the node table.
func (nt *NodeTableBuilder) DynamicProperties(columnName string) *NodeTableBuilder {
	nt.dynamicProperties = NewDynamicProperties(columnName)
	return nt
}

func (nt *NodeTableBuilder) Query() (string, []any) {
	b := nt.Clone()
	b.Ident(nt.tableName)
	if nt.alias != "" {
		b.WriteString(" AS ").Ident(nt.alias)
	}

	b.Indent()
	if nt.key != nil {
		b.NewLine().WriteString("KEY ")
		b.Wrap(func(b *sql.Builder) {
			b.IdentComma(nt.key.Columns...)
		})
	}

	// Handle labels
	if len(nt.labels) == 0 {
		b.NewLine().WriteString("DEFAULT LABEL PROPERTIES ARE ALL COLUMNS")
	} else {
		for _, label := range nt.labels {
			if label.IsDefault {
				b.NewLine().WriteString("DEFAULT LABEL")
			} else if label.Name != "" {
				b.NewLine().WriteString("LABEL ").Ident(label.Name)
			} else {
				continue // Skip empty named labels that aren't default
			}

			// Handle properties for this label
			nt.writePropertiesClause(&b, label.Properties)
		}
	}

	// Handle dynamic label
	if nt.dynamicLabel != nil {
		b.NewLine().WriteString("DYNAMIC LABEL ")
		b.Wrap(func(b *sql.Builder) {
			b.Ident(nt.dynamicLabel.ColumnName)
		})
	}

	// Handle dynamic properties
	if nt.dynamicProperties != nil {
		b.NewLine().WriteString("DYNAMIC PROPERTIES ")
		b.Wrap(func(b *sql.Builder) {
			b.Ident(nt.dynamicProperties.ColumnName)
		})
	}
	b.Dedent()

	return b.String(), b.GetArgs()
}

// writePropertiesClause writes the properties clause for a label.
func (nt *NodeTableBuilder) writePropertiesClause(b *sql.Builder, props *PropertiesConfig) {
	if props == nil {
		b.WriteString(" PROPERTIES ARE ALL COLUMNS")
		return
	}

	switch props.Type {
	case PropertiesTypeNone:
		b.WriteString(" NO PROPERTIES")
	case PropertiesTypeAll:
		if len(props.ExceptColumns) == 0 {
			b.WriteString(" PROPERTIES ARE ALL COLUMNS")
		} else {
			b.WriteString(" PROPERTIES ARE ALL COLUMNS EXCEPT ")
			b.Wrap(func(b *sql.Builder) {
				b.IdentComma(props.ExceptColumns...)
			})
		}
	case PropertiesTypeDerived:
		b.WriteString(" PROPERTIES ")
		b.Wrap(func(b *sql.Builder) {
			for i, prop := range props.DerivedProperties {
				if i > 0 {
					b.Comma()
				}
				b.WriteString(prop.Expression)
				if prop.Alias != "" {
					b.WriteString(" AS ").Ident(prop.Alias)
				}
			}
		})
	default:
		b.WriteString(" PROPERTIES ARE ALL COLUMNS")
	}
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

func (nt *NodeTableBuilder) L() *labelExpr {
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
	tableName         string
	alias             string
	key               *ElementKey
	sourceKey         *ReferenceKey
	destinationKey    *ReferenceKey
	labels            []*Label
	dynamicLabel      *DynamicLabel
	dynamicProperties *DynamicProperties
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

// DefaultLabel adds a default label to the edge table.
func (et *EdgeTableBuilder) DefaultLabel() *LabelBuilder {
	label := NewDefaultLabel()
	et.labels = append(et.labels, label)
	return &LabelBuilder{label: label}
}

// DynamicLabel sets the dynamic label for the edge table.
func (et *EdgeTableBuilder) DynamicLabel(columnName string) *EdgeTableBuilder {
	et.dynamicLabel = NewDynamicLabel(columnName)
	return et
}

// DynamicProperties sets the dynamic properties for the edge table.
func (et *EdgeTableBuilder) DynamicProperties(columnName string) *EdgeTableBuilder {
	et.dynamicProperties = NewDynamicProperties(columnName)
	return et
}

// Query returns query representation of the edge table.
func (et *EdgeTableBuilder) Query() (string, []any) {
	b := et.Clone()
	b.Ident(et.tableName)
	if et.alias != "" {
		b.WriteString(" AS ").Ident(et.alias)
	}

	b.Indent()
	if et.key != nil {
		b.NewLine().WriteString("KEY ")
		b.Wrap(func(b *sql.Builder) {
			b.IdentComma(et.key.Columns...)
		})
	}
	if et.sourceKey != nil {
		b.NewLine().WriteString("SOURCE KEY ")
		b.Wrap(func(b *sql.Builder) {
			b.IdentComma(et.sourceKey.Columns...)
		})
		if et.sourceKey.ReferencedTable != "" && len(et.sourceKey.ReferencedColumns) > 0 {
			b.WriteString(" REFERENCES ")
			b.Ident(et.sourceKey.ReferencedTable)
			b.WriteByte(' ')
			b.Wrap(func(b *sql.Builder) {
				b.IdentComma(et.sourceKey.ReferencedColumns...)
			})
		}
	}
	if et.destinationKey != nil {
		b.NewLine().WriteString("DESTINATION KEY ")
		b.Wrap(func(b *sql.Builder) {
			b.IdentComma(et.destinationKey.Columns...)
		})
		if et.destinationKey.ReferencedTable != "" && len(et.destinationKey.ReferencedColumns) > 0 {
			b.WriteString(" REFERENCES ")
			b.Ident(et.destinationKey.ReferencedTable)
			b.WriteByte(' ')
			b.Wrap(func(b *sql.Builder) {
				b.IdentComma(et.destinationKey.ReferencedColumns...)
			})
		}
	}

	// Handle labels
	if len(et.labels) == 0 {
		b.NewLine().WriteString("DEFAULT LABEL PROPERTIES ARE ALL COLUMNS")
	} else {
		for _, label := range et.labels {
			if label.IsDefault {
				b.NewLine().WriteString("DEFAULT LABEL")
			} else if label.Name != "" {
				b.NewLine().WriteString("LABEL ").Ident(label.Name)
			} else {
				continue // Skip empty named labels that aren't default
			}

			// Handle properties for this label
			et.writePropertiesClause(&b, label.Properties)
		}
	}

	// Handle dynamic label
	if et.dynamicLabel != nil {
		b.NewLine().WriteString("DYNAMIC LABEL ")
		b.Wrap(func(b *sql.Builder) {
			b.Ident(et.dynamicLabel.ColumnName)
		})
	}

	// Handle dynamic properties
	if et.dynamicProperties != nil {
		b.NewLine().WriteString("DYNAMIC PROPERTIES ")
		b.Wrap(func(b *sql.Builder) {
			b.Ident(et.dynamicProperties.ColumnName)
		})
	}
	b.Dedent()

	return b.String(), b.GetArgs()
}

// writePropertiesClause writes the properties clause for a label.
func (et *EdgeTableBuilder) writePropertiesClause(b *sql.Builder, props *PropertiesConfig) {
	if props == nil {
		b.WriteString(" PROPERTIES ARE ALL COLUMNS")
		return
	}

	switch props.Type {
	case PropertiesTypeNone:
		b.WriteString(" NO PROPERTIES")
	case PropertiesTypeAll:
		if len(props.ExceptColumns) == 0 {
			b.WriteString(" PROPERTIES ARE ALL COLUMNS")
		} else {
			b.WriteString(" PROPERTIES ARE ALL COLUMNS EXCEPT ")
			b.Wrap(func(b *sql.Builder) {
				b.IdentComma(props.ExceptColumns...)
			})
		}
	case PropertiesTypeDerived:
		b.WriteString(" PROPERTIES ")
		b.Wrap(func(b *sql.Builder) {
			for i, prop := range props.DerivedProperties {
				if i > 0 {
					b.Comma()
				}
				b.WriteString(prop.Expression)
				if prop.Alias != "" {
					b.WriteString(" AS ").Ident(prop.Alias)
				}
			}
		})
	default:
		b.WriteString(" PROPERTIES ARE ALL COLUMNS")
	}
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

func (et *EdgeTableBuilder) L() *labelExpr {
	if et.alias != "" {
		return L(et.alias)
	}
	return L(et.tableName)
}

// LabelBuilder builds label definitions.
type LabelBuilder struct {
	label *Label
}

// NoProperties sets the label to have no properties.
func (lb *LabelBuilder) NoProperties() *LabelBuilder {
	lb.label.Properties = NewNoProperties()
	return lb
}

// AllProperties sets the label to include all columns as properties.
func (lb *LabelBuilder) AllProperties() *LabelBuilder {
	lb.label.Properties = NewAllProperties()
	return lb
}

// AllPropertiesExcept sets the label to include all columns except the specified ones.
func (lb *LabelBuilder) AllPropertiesExcept(columns ...string) *LabelBuilder {
	lb.label.Properties = NewAllPropertiesExcept(columns...)
	return lb
}

// DerivedProperties sets the label to use derived properties.
func (lb *LabelBuilder) DerivedProperties(props ...*DerivedProperty) *LabelBuilder {
	lb.label.Properties = NewDerivedProperties(props...)
	return lb
}

// Properties is a legacy method that creates derived properties from column names.
// Deprecated: Use DerivedProperties, AllProperties, AllPropertiesExcept, or NoProperties instead.
func (lb *LabelBuilder) Properties(properties ...string) *LabelBuilder {
	derivedProps := make([]*DerivedProperty, len(properties))
	for i, prop := range properties {
		derivedProps[i] = NewDerivedProperty(prop)
	}
	lb.label.Properties = NewDerivedProperties(derivedProps...)
	return lb
}
