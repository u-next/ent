package schema

import (
	"context"
	"fmt"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/googleapis/gax-go/v2"
)

// Spanner adapter for property graph DDL generation.
type Spanner struct {
	adminClient
	dialect.Driver
	schema    string
	edition   string
	dbDialect string
	project   string
	instance  string
}

// adminClient defines the interface for Spanner Admin API operations.
type adminClient interface {
	GetInstance(ctx context.Context, req *instancepb.GetInstanceRequest, opts ...gax.CallOption) (*instancepb.Instance, error)
}

// init loads the Spanner edition and dialect information from the database for later use.
// It returns an error if property graphs are not supported.
func (s *Spanner) init(ctx context.Context) error {
	if s.edition != "" && s.dbDialect != "" {
		return nil // already initialized.
	}

	// Get edition using Admin API
	instanceResp, err := s.adminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: fmt.Sprintf("projects/%s/instances/%s", s.project, s.instance),
	})
	if err != nil {
		return fmt.Errorf("spanner: getting instance edition: %w", err)
	}
	s.edition = instanceResp.Edition.String()

	// Query INFORMATION_SCHEMA.DATABASE_OPTIONS to get the database dialect
	query, args := sql.Dialect(dialect.Spanner).
		Select("OPTION_VALUE").From(sql.Table("DATABASE_OPTIONS").Schema("INFORMATION_SCHEMA")).
		Where(sql.EQ("OPTION_NAME", "database_dialect")).Query()

	rows := &sql.Rows{}
	if err := s.Query(ctx, query, args, rows); err != nil {
		return fmt.Errorf("spanner: querying database dialect: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&s.dbDialect); err != nil {
			return fmt.Errorf("spanner: scanning database dialect: %w", err)
		}
	}

	return nil
}

func (s *Spanner) tableExist(ctx context.Context, conn dialect.ExecQuerier, name string) (bool, error) {
	return false, nil
}

func (s *Spanner) graphExist(ctx context.Context, drv dialect.ExecQuerier, name string) (bool, error) {
	// Ensure we're initialized
	if err := s.init(ctx); err != nil {
		return false, err
	}

	// Check if property graphs are supported in this instance/database
	if !s.supportsGraph() {
		return false, fmt.Errorf("property graphs are not supported in %s edition with %s dialect",
			s.edition, s.dbDialect)
	}

	// Query INFORMATION_SCHEMA.PROPERTY_GRAPHS to check if the graph exists
	query, args := sql.Dialect(dialect.Spanner).
		Select(sql.Count("*")).From(sql.Table("PROPERTY_GRAPHS").Schema("INFORMATION_SCHEMA")).
		Where(sql.And(
			s.matchSchema(),
			sql.EQ("PROPERTY_GRAPH_NAME", name),
		)).Query()
	return exist(ctx, drv, query, args...)
}

// matchSchema returns the predicate for matching table schema.
func (s *Spanner) matchSchema(columns ...string) *sql.Predicate {
	column := "PROPERTY_GRAPH_SCHEMA"
	if len(columns) > 0 {
		column = columns[0]
	}
	if s.schema != "" {
		return sql.EQ(column, s.schema)
	}
	// For Spanner, if no schema is set, we don't filter by schema
	// since Spanner doesn't support CURRENT_SCHEMA()
	return sql.P()
}

func (s *Spanner) atOpen(dialect.ExecQuerier) (migrate.Driver, error) {
	return nil, nil
}

func (s *Spanner) atTable(*Table, *schema.Table) {
	return
}

func (s *Spanner) supportsDefault(*Column) bool {
	return false
}

func (s *Spanner) atTypeC(*Column, *schema.Column) error {
	return nil
}

func (s *Spanner) atUniqueC(*Table, *Column, *schema.Table, *schema.Column) {
	return
}

func (s *Spanner) atIncrementC(*schema.Table, *schema.Column) {
	return
}

func (s *Spanner) atIncrementT(*schema.Table, int64) {
	return
}

func (s *Spanner) atIndex(*Index, *schema.Table, *schema.Index) error {
	return nil
}

func (s *Spanner) atTypeRangeSQL(t ...string) string {
	return ""
}

// supportsGraph checks if the current Spanner instance/database supports property graphs.
func (s *Spanner) supportsGraph() bool {
	return (s.edition == "ENTERPRISE" || s.edition == "ENTERPRISE_PLUS") &&
		s.dbDialect == "GOOGLE_STANDARD_SQL"
}

// FIXME: refactor to other file
// generateSpannerPropertyGraphDDL generates Spanner-specific property graph DDL.
func generateSpannerPropertyGraphDDL(ctx context.Context, args DDLArgs) (string, error) {
	pg := args.PropertyGraphs[0]
	var stmt strings.Builder

	stmt.WriteString("CREATE PROPERTY GRAPH ")

	// Add schema prefix if specified
	if pg.Schema != "" {
		stmt.WriteString(pg.Schema)
		stmt.WriteString(".")
	}

	stmt.WriteString(pg.Name)

	// Add node tables
	if len(pg.NodeTables) > 0 {
		stmt.WriteString("\n  NODE TABLES (")
		for i, nodeTable := range pg.NodeTables {
			if i > 0 {
				stmt.WriteString(",")
			}
			stmt.WriteString("\n    ")
			writeSpannerNodeTable(&stmt, nodeTable)
		}
		stmt.WriteString("\n  )")
	}

	// Add edge tables
	if len(pg.EdgeTables) > 0 {
		stmt.WriteString("\n  EDGE TABLES (")
		for i, edgeTable := range pg.EdgeTables {
			if i > 0 {
				stmt.WriteString(",")
			}
			stmt.WriteString("\n    ")
			writeSpannerEdgeTable(&stmt, edgeTable)
		}
		stmt.WriteString("\n  )")
	}

	stmt.WriteString(";")
	return stmt.String(), nil
}

// writeSpannerNodeTable writes a node table definition to the statement builder for Spanner.
func writeSpannerNodeTable(stmt *strings.Builder, nodeTable *NodeTable) {
	// For property graphs, just output the table name (simplified format)
	stmt.WriteString(nodeTable.TableName)
	for _, label := range nodeTable.Labels {
		if label.IsDefault {
			stmt.WriteString("\n      DEFAULT LABEL")
		} else {
			stmt.WriteString("\n      LABEL ")
			stmt.WriteString(label.Name)
		}

		switch label.Properties.Type {
		case PropertiesTypeNone:
			stmt.WriteString("\n      NO PROPERTIES")
		case PropertiesTypeAll:
			stmt.WriteString("\n      PROPERTIES ARE ALL COLUMNS")
			if len(label.Properties.ExcludedColumns) > 0 {
				stmt.WriteString(" EXCEPT (")
				for i, col := range label.Properties.ExcludedColumns {
					if i > 0 {
						stmt.WriteString(", ")
					}
					stmt.WriteString(col)
				}
				stmt.WriteString(")")
			}
		case PropertiesTypeDerived:
			stmt.WriteString("\n      PROPERTIES (")
			for i, prop := range label.Properties.DerivedProperties {
				if i > 0 {
					stmt.WriteString(", ")
				}
				stmt.WriteString(prop.Expression)
				if prop.Alias != "" {
					stmt.WriteString(" AS ")
					stmt.WriteString(prop.Alias)
				}
			}
			stmt.WriteString(")")
		}
	}
}

// writeSpannerEdgeTable writes an edge table definition to the statement builder for Spanner.
func writeSpannerEdgeTable(stmt *strings.Builder, edgeTable *EdgeTable) {
	stmt.WriteString(edgeTable.TableName)

	// Add source key with proper formatting
	if edgeTable.SourceKey != nil {
		stmt.WriteString("\n      SOURCE KEY (")
		for i, col := range edgeTable.SourceKey.Columns {
			if i > 0 {
				stmt.WriteString(", ")
			}
			stmt.WriteString(col)
		}
		stmt.WriteString(")")

		if edgeTable.SourceKey.ReferencedTable != "" && len(edgeTable.SourceKey.ReferencedColumns) > 0 {
			stmt.WriteString(" REFERENCES ")
			stmt.WriteString(edgeTable.SourceKey.ReferencedTable)
			stmt.WriteString(" (")
			for i, col := range edgeTable.SourceKey.ReferencedColumns {
				if i > 0 {
					stmt.WriteString(", ")
				}
				stmt.WriteString(col)
			}
			stmt.WriteString(")")
		}
	}

	// Add destination key with proper formatting
	if edgeTable.DestinationKey != nil {
		stmt.WriteString("\n      DESTINATION KEY (")
		for i, col := range edgeTable.DestinationKey.Columns {
			if i > 0 {
				stmt.WriteString(", ")
			}
			stmt.WriteString(col)
		}
		stmt.WriteString(")")

		if edgeTable.DestinationKey.ReferencedTable != "" && len(edgeTable.DestinationKey.ReferencedColumns) > 0 {
			stmt.WriteString(" REFERENCES ")
			stmt.WriteString(edgeTable.DestinationKey.ReferencedTable)
			stmt.WriteString(" (")
			for i, col := range edgeTable.DestinationKey.ReferencedColumns {
				if i > 0 {
					stmt.WriteString(", ")
				}
				stmt.WriteString(col)
			}
			stmt.WriteString(")")
		}
	}

	for _, edgeLabel := range edgeTable.Labels {
		if edgeLabel.IsDefault {
			stmt.WriteString("\n      DEFAULT LABEL")
		} else {
			stmt.WriteString("\n      LABEL ")
			stmt.WriteString(edgeLabel.Name)
		}

		switch edgeLabel.Properties.Type {
		case PropertiesTypeNone:
			stmt.WriteString("\n      NO PROPERTIES")
		case PropertiesTypeAll:
			stmt.WriteString("\n      PROPERTIES ARE ALL COLUMNS")
			if len(edgeLabel.Properties.ExcludedColumns) > 0 {
				stmt.WriteString(" EXCEPT (")
				for i, col := range edgeLabel.Properties.ExcludedColumns {
					if i > 0 {
						stmt.WriteString(", ")
					}
					stmt.WriteString(col)
				}
				stmt.WriteString(")")
			}
		case PropertiesTypeDerived:
			stmt.WriteString("\n      PROPERTIES (")
			for i, prop := range edgeLabel.Properties.DerivedProperties {
				if i > 0 {
					stmt.WriteString(", ")
				}
				stmt.WriteString(prop.Expression)
				if prop.Alias != "" {
					stmt.WriteString(" AS ")
					stmt.WriteString(prop.Alias)
				}
			}
			stmt.WriteString(")")
		}
	}
}
