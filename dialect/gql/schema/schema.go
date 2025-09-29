// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Package schema contains all property graph schema logic for GQL dialects (Spanner).
package schema

import (
	"context"
	"fmt"
	"slices"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
)

// PropertyGraph represents a property graph schema definition for GQL dialects.
type PropertyGraph struct {
	schema.Object
	Name       string
	Schema     string // Database schema containing the property graph
	NodeTables []*NodeTable
	EdgeTables []*EdgeTable
	Annotation *entsql.Annotation
	Comment    string
	Options    []string // Property graph options following Spanner HINT syntax
	Pos        string   // filename:line of the ent schema definition
}

// NewPropertyGraph returns a new property graph with the given name.
func NewPropertyGraph(name string) *PropertyGraph {
	return &PropertyGraph{
		Name: name,
	}
}

// SetComment sets the property graph comment.
func (pg *PropertyGraph) SetComment(c string) *PropertyGraph {
	pg.Comment = c
	return pg
}

// SetSchema sets the property graph schema.
func (pg *PropertyGraph) SetSchema(s string) *PropertyGraph {
	pg.Schema = s
	return pg
}

// SetPos sets the property graph position.
func (pg *PropertyGraph) SetPos(p string) *PropertyGraph {
	pg.Pos = p
	return pg
}

// AddNodeTable adds a node table to the property graph.
func (pg *PropertyGraph) AddNodeTable(nt *NodeTable) *PropertyGraph {
	pg.NodeTables = append(pg.NodeTables, nt)
	return pg
}

// AddEdgeTable adds an edge table to the property graph.
func (pg *PropertyGraph) AddEdgeTable(et *EdgeTable) *PropertyGraph {
	pg.EdgeTables = append(pg.EdgeTables, et)
	return pg
}

// SetAnnotation sets the entsql.Annotation on the property graph.
func (pg *PropertyGraph) SetAnnotation(ant *entsql.Annotation) *PropertyGraph {
	pg.Annotation = ant
	return pg
}

// AddOptions adds options to the property graph.
func (pg *PropertyGraph) AddOptions(opts ...string) *PropertyGraph {
	pg.Options = append(pg.Options, opts...)
	return pg
}

// NodeTable represents a node table definition in a property graph.
type NodeTable struct {
	TableName         string             // Name of the underlying table
	Alias             string             // Optional alias for the table
	Key               *ElementKey        // Element key definition
	Labels            []*Label           // Labels for this node type
	DynamicLabel      *DynamicLabel      // Dynamic label definition
	DynamicProperties *DynamicProperties // Dynamic properties definition
}

// NewNodeTable returns a new node table with the given table name.
func NewNodeTable(tableName string) *NodeTable {
	return &NodeTable{
		TableName: tableName,
	}
}

// SetAlias sets the node table alias.
func (nt *NodeTable) SetAlias(alias string) *NodeTable {
	nt.Alias = alias
	return nt
}

// SetKey sets the element key for the node table.
func (nt *NodeTable) SetKey(key *ElementKey) *NodeTable {
	nt.Key = key
	return nt
}

// AddLabel adds a label to the node table.
func (nt *NodeTable) AddLabel(label *Label) *NodeTable {
	nt.Labels = append(nt.Labels, label)
	return nt
}

// SetDynamicLabel sets the dynamic label definition.
func (nt *NodeTable) SetDynamicLabel(dl *DynamicLabel) *NodeTable {
	nt.DynamicLabel = dl
	return nt
}

// SetDynamicProperties sets the dynamic properties definition.
func (nt *NodeTable) SetDynamicProperties(dp *DynamicProperties) *NodeTable {
	nt.DynamicProperties = dp
	return nt
}

// EdgeTable represents an edge table definition in a property graph.
type EdgeTable struct {
	TableName         string             // Name of the underlying table
	Alias             string             // Optional alias for the table
	Key               *ElementKey        // Element key definition
	SourceKey         *ReferenceKey      // Source key definition
	DestinationKey    *ReferenceKey      // Destination key definition
	Labels            []*Label           // Labels for this edge type
	DynamicLabel      *DynamicLabel      // Dynamic label definition
	DynamicProperties *DynamicProperties // Dynamic properties definition
}

// NewEdgeTable returns a new edge table with the given table name.
func NewEdgeTable(tableName string) *EdgeTable {
	return &EdgeTable{
		TableName: tableName,
	}
}

// SetAlias sets the edge table alias.
func (et *EdgeTable) SetAlias(alias string) *EdgeTable {
	et.Alias = alias
	return et
}

// SetKey sets the element key for the edge table.
func (et *EdgeTable) SetKey(key *ElementKey) *EdgeTable {
	et.Key = key
	return et
}

// SetSourceKey sets the source key for the edge table.
func (et *EdgeTable) SetSourceKey(key *ReferenceKey) *EdgeTable {
	et.SourceKey = key
	return et
}

// SetDestinationKey sets the destination key for the edge table.
func (et *EdgeTable) SetDestinationKey(key *ReferenceKey) *EdgeTable {
	et.DestinationKey = key
	return et
}

// AddLabel adds a label to the edge table.
func (et *EdgeTable) AddLabel(label *Label) *EdgeTable {
	et.Labels = append(et.Labels, label)
	return et
}

// SetDynamicLabel sets the dynamic label definition.
func (et *EdgeTable) SetDynamicLabel(dl *DynamicLabel) *EdgeTable {
	et.DynamicLabel = dl
	return et
}

// SetDynamicProperties sets the dynamic properties definition.
func (et *EdgeTable) SetDynamicProperties(dp *DynamicProperties) *EdgeTable {
	et.DynamicProperties = dp
	return et
}

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
	Name       string      // Label name
	IsDefault  bool        // Whether this is the default label
	Properties *Properties // Properties associated with this label
}

// NewLabel returns a new label with the given name.
func NewLabel(name string) *Label {
	return &Label{
		Name: name,
	}
}

func NewDefaultLabel() *Label {
	return &Label{
		IsDefault: true,
	}
}

// SetDefault marks this label as the default label.
func (l *Label) SetDefault() *Label {
	l.IsDefault = true
	return l
}

// SetProperties sets the properties for this label.
func (l *Label) SetProperties(props *Properties) *Label {
	l.Properties = props
	return l
}

// Properties represents a properties definition.
type Properties struct {
	Type              PropertiesType     // Type of properties definition
	AllColumns        bool               // Whether to include all columns
	ExcludedColumns   []string           // Columns to exclude when using all columns
	DerivedProperties []*DerivedProperty // List of derived properties
}

// PropertiesType represents the type of properties definition.
type PropertiesType int

const (
	PropertiesTypeNone PropertiesType = iota
	PropertiesTypeAll
	PropertiesTypeDerived
)

// NewNoProperties returns a properties definition with no properties.
func NewNoProperties() *Properties {
	return &Properties{
		Type: PropertiesTypeNone,
	}
}

// NewAllProperties returns a properties definition that includes all columns.
func NewAllProperties() *Properties {
	return &Properties{
		Type:       PropertiesTypeAll,
		AllColumns: true,
	}
}

// NewAllPropertiesExcept returns a properties definition that includes all columns except the specified ones.
func NewAllPropertiesExcept(excludedColumns ...string) *Properties {
	return &Properties{
		Type:            PropertiesTypeAll,
		AllColumns:      true,
		ExcludedColumns: excludedColumns,
	}
}

// NewDerivedProperties returns a properties definition with derived properties.
func NewDerivedProperties(props ...*DerivedProperty) *Properties {
	return &Properties{
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

// SetAlias sets the alias for the derived property.
func (dp *DerivedProperty) SetAlias(alias string) *DerivedProperty {
	dp.Alias = alias
	return dp
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

type driver struct {
	gqlDialect
	schema.Differ
	migrate.PlanApplier
}

var drivers = func(v string) map[string]driver {
	return map[string]driver{
		dialect.Spanner: {
			&Spanner{
				Driver: nopDriver{dialect: dialect.Spanner},
			},
			schema.Differ(nil), // No differ implementation yet
			migrate.PlanApplier(nil),
		},
	}
}

// DDLArgs contains arguments for property graph DDL generation.
type DDLArgs struct {
	// Dialect and Version of the target database.
	Dialect, Version string
	// PropertyGraph to generate DDL for.
	PropertyGraphs []*PropertyGraph
	// Drop indicates whether to generate DROP instead of CREATE.
	Drop bool
	// IfExists adds IF EXISTS clause for DROP operations.
	IfExists bool
	// HashSymbols indicates whether to hash long symbols in the DDL.
	HashSymbols bool
	// Options for the migration plan.
	Options []migrate.PlanOption
}

// Deprecated: use DDL instead.
func PropertyGraphDDL(ctx context.Context, args DDLArgs) (string, error) {
	switch args.Dialect {
	case dialect.Spanner:
	default:
		return "", fmt.Errorf("unsupported dialect %q", args.Dialect)
	}
	return generateSpannerPropertyGraphDDL(ctx, args)
}

// DDL generates CREATE or DROP PROPERTY GRAPH DDL statement.
func DDL(ctx context.Context, args DDLArgs) (string, error) {
	// TODO: Handle DDLArgs Drop, IfExists, multiple PropertyGraphs, and Options.
	args.Options = append([]migrate.PlanOption{func(o *migrate.PlanOptions) {
		o.Mode = migrate.PlanModeDump
		o.Indent = "  "
	}}, args.Options...)
	d, ok := drivers(args.Version)[args.Dialect]
	if !ok {
		return "", fmt.Errorf("unsupported dialect %q", args.Dialect)
	}

	var err error

	m := &Migrate{
		gqlDialect:  d,
		dialect:     args.Dialect,
		hashSymbols: args.HashSymbols,
	}
	// TODO: read current state from database.
	r, err := m.StateReader(args.PropertyGraphs...).ReadState(ctx)
	if err != nil {
		return "", err
	}
	var c schema.Changes
	if slices.ContainsFunc(args.PropertyGraphs, func(t *PropertyGraph) bool { return t.Schema != "" }) {
		// TODO: implement RealmDiff to support multiple schemas.
		c, err = d.RealmDiff(&schema.Realm{}, r)
	} else {
		// TODO: implement SchemaDiff to support property graphs.
		c, err = d.SchemaDiff(&schema.Schema{}, r.Schemas[0])
	}
	if err != nil {
		return "", err
	}
	p, err := d.PlanChanges(ctx, "dump", c, args.Options...)
	if err != nil {
		return "", err
	}

	for _, pg := range args.PropertyGraphs {
		p.Directives = append(p.Directives, fmt.Sprintf(
			"-- atlas:pos %s%s[type=%s]",
			func() string {
				if pg.Schema != "" {
					return pg.Schema + "[type=schema]."
				}
				return ""
			}(),
			pg.Name,
			pg.Pos,
		))
	}
	f, err := migrate.DefaultFormatter.FormatFile(p)
	if err != nil {
		return "", err
	}
	return string(f.Bytes()), nil
}
