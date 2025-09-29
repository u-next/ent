// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	sqlschema "entgo.io/ent/dialect/sql/schema"
)

// gqlDialect interface for internal use with property graph operations.
type gqlDialect interface {
	dialect.Driver
	init(context.Context) error
	graphExist(context.Context, dialect.ExecQuerier, string) (bool, error)
}

// Migrate manages property graph definitions in a state-based approach.
// Unlike traditional SQL migrations, property graphs are immutable and replaced as a whole.
type Migrate struct {
	gqlDialect
	driver      dialect.Driver
	dialect     string
	schema      string
	stateTable  string
	inspector   Inspector
	hashSymbols bool
}

// MigrateOption configures the property graph migrator.
type MigrateOption func(*Migrate)

// WithSchema sets the database schema for property graphs.
func WithSchema(schema string) MigrateOption {
	return func(m *Migrate) {
		m.schema = schema
	}
}

// WithStateTable sets the table name for tracking applied definitions.
func WithStateTable(table string) MigrateOption {
	return func(m *Migrate) {
		m.stateTable = table
	}
}

// NewMigrate creates a new property graph migrator.
func NewMigrate(driver dialect.Driver, opts ...MigrateOption) (*Migrate, error) {
	m := &Migrate{
		driver:     driver,
		dialect:    driver.Dialect(),
		stateTable: "_property_graph_state",
	}

	for _, opt := range opts {
		opt(m)
	}

	// Initialize inspector
	inspector, err := NewInspector(m.dialect, driver)
	if err != nil {
		return nil, fmt.Errorf("creating inspector: %w", err)
	}
	m.inspector = inspector

	return m, nil
}

// Apply reads property graph definitions and ensures database state matches.
// This is the main operation - compares current DB state with definitions and replaces if needed.
func (m *Migrate) Apply(ctx context.Context, definitions []*PropertyGraph) error {
	if err := m.initStateTable(ctx); err != nil {
		return fmt.Errorf("initializing state table: %w", err)
	}

	// Inspect current database state
	current, err := m.inspector.InspectPropertyGraphs(ctx, m.schema)
	if err != nil {
		return fmt.Errorf("inspecting current property graphs: %w", err)
	}

	// Convert definitions to desired state
	desired := make(map[string]*PropertyGraph)
	for _, pg := range definitions {
		desired[pg.Name] = pg
	}

	// Execute in transaction
	tx, err := m.driver.Tx(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	var operations []Operation

	// Drop graphs that are no longer defined
	for name, currentPG := range current.Graphs {
		if _, exists := desired[name]; !exists {
			if err := m.dropPropertyGraph(ctx, tx, currentPG); err != nil {
				return fmt.Errorf("dropping property graph %s: %w", name, err)
			}
			operations = append(operations, Operation{
				Type:      "DROP",
				GraphName: name,
				AppliedAt: time.Now(),
			})
		}
	}

	// Create or replace graphs as needed
	for name, desiredPG := range desired {
		currentPG, exists := current.Graphs[name]

		if !exists {
			// Create new graph
			if err := m.createPropertyGraph(ctx, tx, desiredPG); err != nil {
				return fmt.Errorf("creating property graph %s: %w", name, err)
			}
			operations = append(operations, Operation{
				Type:      "CREATE",
				GraphName: name,
				AppliedAt: time.Now(),
			})
		} else if m.graphChanged(currentPG, desiredPG) {
			// Replace existing graph
			if err := m.replacePropertyGraph(ctx, tx, desiredPG); err != nil {
				return fmt.Errorf("replacing property graph %s: %w", name, err)
			}
			operations = append(operations, Operation{
				Type:      "REPLACE",
				GraphName: name,
				AppliedAt: time.Now(),
			})
		}
		// If graph exists and hasn't changed, do nothing
	}

	// Record operations
	for _, op := range operations {
		if err := m.recordOperation(ctx, tx, op); err != nil {
			return fmt.Errorf("recording operation: %w", err)
		}
	}

	return tx.Commit()
}

// StateReader returns an atlas migrate.StateReader returning the state as described by the Ent table slice.
func (m *Migrate) StateReader(graphs ...*PropertyGraph) migrate.StateReaderFunc {
	return func(ctx context.Context) (*schema.Realm, error) {
		if m.gqlDialect == nil {
			drv, err := m.entDialect(ctx, m.driver)
			if err != nil {
				return nil, err
			}
			m.gqlDialect = drv
		}
		if m.hashSymbols {
			m.setupGraphs(graphs)
		}
		return m.realm(graphs)
	}
}

// realm constructs a schema.Realm from the given property graphs.
func (m *Migrate) realm(graphs []*PropertyGraph) (*schema.Realm, error) {
	return &schema.Realm{}, nil
}

// entDialect returns the Ent dialect as configured by the dialect option.
func (a *Migrate) entDialect(ctx context.Context, drv dialect.Driver) (gqlDialect, error) {
	var d gqlDialect
	switch a.dialect {
	case dialect.Spanner:
		d = &Spanner{Driver: drv}
	default:
		return nil, fmt.Errorf("sql/schema: unsupported dialect %q", a.dialect)
	}
	if err := d.init(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

func (m *Migrate) setupGraphs(graphs []*PropertyGraph) {
	// TODO: hash long symbols in the property graph
}

// Status returns the current status of property graphs.
func (m *Migrate) Status(ctx context.Context) (*Status, error) {
	// Inspect current state
	current, err := m.inspector.InspectPropertyGraphs(ctx, m.schema)
	if err != nil {
		return nil, fmt.Errorf("inspecting property graphs: %w", err)
	}

	// Get last operations
	operations, err := m.getRecentOperations(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("getting recent operations: %w", err)
	}

	// Get last applied time
	var lastApplied *time.Time
	if len(operations) != 0 {
		lastApplied = &operations[0].AppliedAt
	}

	return &Status{
		Schema:           m.schema,
		PropertyGraphs:   current.Graphs,
		RecentOperations: operations,
		LastApplied:      lastApplied,
	}, nil
}

// Inspector interface for examining database state
type Inspector interface {
	InspectPropertyGraphs(ctx context.Context, schema string) (*PropertyGraphRealm, error)
	InspectPropertyGraph(ctx context.Context, schema, name string) (*PropertyGraph, error)
}

// NewInspector creates an inspector for the given dialect.
func NewInspector(dialect string, driver dialect.Driver) (Inspector, error) {
	switch dialect {
	case "spanner":
		return &SpannerInspector{driver: driver}, nil
	default:
		return &GenericInspector{driver: driver}, nil
	}
}

// SpannerInspector inspects Spanner property graphs using INFORMATION_SCHEMA.
type SpannerInspector struct {
	driver dialect.Driver
}

func (i *SpannerInspector) InspectPropertyGraphs(ctx context.Context, schema string) (*PropertyGraphRealm, error) {
	realm := &PropertyGraphRealm{
		Schema: schema,
		Graphs: make(map[string]*PropertyGraph),
	}

	query := `SELECT graph_name FROM INFORMATION_SCHEMA.PROPERTY_GRAPHS`
	var args []any
	if schema != "" {
		query += ` WHERE schema_name = ?`
		args = append(args, schema)
	}

	rows := &sql.Rows{}
	if err := i.driver.Query(ctx, query, args, rows); err != nil {
		// If query fails, return empty realm (property graphs may not be supported)
		return realm, nil
	}
	defer rows.Close()

	for rows.Next() {
		var graphName string
		if err := rows.Scan(&graphName); err != nil {
			return nil, err
		}

		pg, err := i.InspectPropertyGraph(ctx, schema, graphName)
		if err != nil {
			return nil, fmt.Errorf("inspecting property graph %s: %w", graphName, err)
		}
		realm.Graphs[graphName] = pg
	}

	return realm, nil
}

func (i *SpannerInspector) InspectPropertyGraph(ctx context.Context, schema, name string) (*PropertyGraph, error) {
	pg := NewPropertyGraph(name)
	return pg, nil
}

// GenericInspector provides basic inspection for non-Spanner dialects.
type GenericInspector struct {
	driver dialect.Driver
}

func (i *GenericInspector) InspectPropertyGraphs(ctx context.Context, schema string) (*PropertyGraphRealm, error) {
	return &PropertyGraphRealm{
		Schema: schema,
		Graphs: make(map[string]*PropertyGraph),
	}, nil
}

func (i *GenericInspector) InspectPropertyGraph(ctx context.Context, schema, name string) (*PropertyGraph, error) {
	return nil, fmt.Errorf("property graph inspection not supported for this dialect")
}

// State management types

// PropertyGraphRealm represents the current state of property graphs in the database.
type PropertyGraphRealm struct {
	Schema string                    `json:"schema"`
	Graphs map[string]*PropertyGraph `json:"graphs"`
}

// Operation represents a single operation performed on property graphs.
type Operation struct {
	Type      string    `json:"type"` // CREATE, DROP, REPLACE
	GraphName string    `json:"graph_name"`
	AppliedAt time.Time `json:"applied_at"`
}

// Status represents the current status of property graphs.
type Status struct {
	Schema           string                    `json:"schema"`
	PropertyGraphs   map[string]*PropertyGraph `json:"property_graphs"`
	RecentOperations []Operation               `json:"recent_operations"`
	LastApplied      *time.Time                `json:"last_applied,omitempty"`
}

// Implementation methods

func (m *Migrate) createPropertyGraph(ctx context.Context, conn dialect.ExecQuerier, pg *PropertyGraph) error {
	ddl, err := DDL(ctx, DDLArgs{
		Dialect:        m.dialect,
		PropertyGraphs: []*PropertyGraph{pg},
		Drop:           false,
		IfExists:       false,
	})
	if err != nil {
		return err
	}
	return conn.Exec(ctx, ddl, nil, nil)
}

func (m *Migrate) replacePropertyGraph(ctx context.Context, conn dialect.ExecQuerier, pg *PropertyGraph) error {
	ddl, err := DDL(ctx, DDLArgs{
		Dialect:        m.dialect,
		PropertyGraphs: []*PropertyGraph{pg},
		Drop:           false,
		IfExists:       false,
	})
	if err != nil {
		return err
	}

	// Convert to CREATE OR REPLACE
	replaceSQL := strings.Replace(ddl, "CREATE PROPERTY GRAPH", "CREATE OR REPLACE PROPERTY GRAPH", 1)
	if replaceSQL == ddl {
		return fmt.Errorf("failed to generate CREATE OR REPLACE statement")
	}

	return conn.Exec(ctx, replaceSQL, nil, nil)
}

func (m *Migrate) dropPropertyGraph(ctx context.Context, conn dialect.ExecQuerier, pg *PropertyGraph) error {
	ddl, err := DDL(ctx, DDLArgs{
		Dialect:        m.dialect,
		PropertyGraphs: []*PropertyGraph{pg},
		Drop:           true,
		IfExists:       true,
	})
	if err != nil {
		return err
	}
	return conn.Exec(ctx, ddl, nil, nil)
}

func (m *Migrate) graphChanged(current, desired *PropertyGraph) bool {
	// Simple comparison - property graphs are immutable, so any structural change requires replacement
	// In practice, you might want to generate DDL for both and compare the SQL strings
	if current.Comment != desired.Comment {
		return true
	}
	if len(current.NodeTables) != len(desired.NodeTables) {
		return true
	}
	if len(current.EdgeTables) != len(desired.EdgeTables) {
		return true
	}
	// For a more robust comparison, you could generate the DDL for both and compare
	return false
}

func (m *Migrate) initStateTable(ctx context.Context) error {
	sql := m.generateStateTableSQL()
	return m.driver.Exec(ctx, sql, nil, nil)
}

func (m *Migrate) generateStateTableSQL() string {
	switch m.dialect {
	case "spanner":
		return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id STRING(255) NOT NULL,
	operation_type STRING(20) NOT NULL,
	graph_name STRING(255) NOT NULL,
	applied_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (id)`, m.stateTable)
	default:
		return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id VARCHAR(255) PRIMARY KEY,
	operation_type VARCHAR(20) NOT NULL,
	graph_name VARCHAR(255) NOT NULL,
	applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, m.stateTable)
	}
}

func (m *Migrate) recordOperation(ctx context.Context, conn dialect.ExecQuerier, op Operation) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, operation_type, graph_name, applied_at)
		VALUES (?, ?, ?, ?)
	`, m.stateTable)

	args := []any{
		fmt.Sprintf("%d", time.Now().UnixNano()),
		op.Type,
		op.GraphName,
		op.AppliedAt,
	}

	return conn.Exec(ctx, query, args, nil)
}

func (m *Migrate) getRecentOperations(ctx context.Context, limit int) ([]Operation, error) {
	query := fmt.Sprintf(`
		SELECT operation_type, graph_name, applied_at
		FROM %s
		ORDER BY applied_at DESC
		LIMIT ?
	`, m.stateTable)

	rows := &sql.Rows{}
	if err := m.driver.Query(ctx, query, []any{limit}, rows); err != nil {
		return nil, err
	}
	defer rows.Close()

	var operations []Operation
	for rows.Next() {
		var op Operation
		if err := rows.Scan(&op.Type, &op.GraphName, &op.AppliedAt); err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}

	return operations, nil
}

// exist checks if the given COUNT query returns a value >= 1.
func exist(ctx context.Context, conn dialect.ExecQuerier, query string, args ...any) (bool, error) {
	rows := &sql.Rows{}
	if err := conn.Query(ctx, query, args, rows); err != nil {
		return false, fmt.Errorf("reading schema information %w", err)
	}
	defer rows.Close()
	n, err := sql.ScanInt(rows)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// PropertyGraphBuilder helps construct property graphs from SQL table schemas.
type PropertyGraphBuilder struct {
	pg *PropertyGraph
}

// NewPropertyGraphBuilder creates a new property graph builder.
func NewPropertyGraphBuilder(name string) *PropertyGraphBuilder {
	return &PropertyGraphBuilder{
		pg: NewPropertyGraph(name),
	}
}

// AddNodeTableFromSQLTable creates a node table from an SQL table schema.
func (b *PropertyGraphBuilder) AddNodeTableFromSQLTable(table *sqlschema.Table, options ...NodeTableOption) *PropertyGraphBuilder {
	nt := NewNodeTable(table.Name)

	// Apply options
	for _, opt := range options {
		opt(nt)
	}

	// If no explicit key is set, use the primary key columns
	if nt.Key == nil && len(table.PrimaryKey) > 0 {
		var keyColumns []string
		for _, col := range table.PrimaryKey {
			keyColumns = append(keyColumns, col.Name)
		}
		nt.SetKey(NewElementKey(keyColumns...))
	}

	// If no properties are explicitly set, default to all columns
	// TODO: handle labels and properties more granularly
	nt.AddLabel(NewDefaultLabel().SetProperties(NewAllProperties()))

	b.pg.AddNodeTable(nt)
	return b
}

// AddEdgeTableFromSQLTable creates an edge table from an SQL table schema.
func (b *PropertyGraphBuilder) AddEdgeTableFromSQLTable(table *sqlschema.Table, options ...EdgeTableOption) *PropertyGraphBuilder {
	et := NewEdgeTable(table.Name)

	// Apply options
	for _, opt := range options {
		opt(et)
	}

	// If no explicit key is set, use the primary key columns
	if et.Key == nil && len(table.PrimaryKey) > 0 {
		var keyColumns []string
		for _, col := range table.PrimaryKey {
			keyColumns = append(keyColumns, col.Name)
		}
		et.SetKey(NewElementKey(keyColumns...))
	}

	// If no properties are explicitly set, default to all columns
	// TODO: handle labels and properties more granularly
	et.AddLabel(NewDefaultLabel().SetProperties(NewAllProperties()))

	b.pg.AddEdgeTable(et)
	return b
}

// Build returns the constructed property graph.
func (b *PropertyGraphBuilder) Build() *PropertyGraph {
	return b.pg
}

// NodeTableOption is a function that configures a node table.
type NodeTableOption func(*NodeTable)

// WithNodeKey sets the element key for the node table.
func WithNodeKey(columns ...string) NodeTableOption {
	return func(nt *NodeTable) {
		nt.SetKey(NewElementKey(columns...))
	}
}

// WithNodeProperties sets the properties for the node table.
func WithNodeProperties(props *Properties) NodeTableOption {
	return func(nt *NodeTable) {
		nt.AddLabel(NewDefaultLabel().SetProperties(props))
	}
}

// EdgeTableOption is a function that configures an edge table.
type EdgeTableOption func(*EdgeTable)

// WithEdgeKey sets the element key for the edge table.
func WithEdgeKey(columns ...string) EdgeTableOption {
	return func(et *EdgeTable) {
		et.SetKey(NewElementKey(columns...))
	}
}

// WithSourceKey sets the source key for the edge table.
func WithSourceKey(columns []string, refTable string, refColumns []string) EdgeTableOption {
	return func(et *EdgeTable) {
		et.SetSourceKey(NewReferenceKey(columns, refTable, refColumns))
	}
}

// WithDestinationKey sets the destination key for the edge table.
func WithDestinationKey(columns []string, refTable string, refColumns []string) EdgeTableOption {
	return func(et *EdgeTable) {
		et.SetDestinationKey(NewReferenceKey(columns, refTable, refColumns))
	}
}

// WithEdgeProperties sets the properties for the edge table.
func WithEdgeProperties(props *Properties) EdgeTableOption {
	return func(et *EdgeTable) {
		et.AddLabel(NewDefaultLabel().SetProperties(props))
	}
}
