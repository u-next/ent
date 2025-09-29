// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
)

// Spanner adapter for Atlas migration engine.
type Spanner struct {
	dialect.Driver
	schema  string
	version string
}

// init loads the Spanner version from the database for later use in the migration process.
func (d *Spanner) init(ctx context.Context) error {
	if d.version != "" {
		return nil // already initialized.
	}
	// Spanner doesn't have a version query like MySQL/Postgres
	// We'll set a default version for compatibility
	d.version = "1.0.0"
	return nil
}

// tableExist checks if a table exists in the current schema.
func (d *Spanner) tableExist(ctx context.Context, conn dialect.ExecQuerier, name string) (bool, error) {
	// Spanner uses INFORMATION_SCHEMA (case sensitive) and doesn't use schema/database names
	// like MySQL or PostgreSQL - tables exist at the database level
	query := "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_name = ?"
	rows := &sql.Rows{}
	if err := conn.Query(ctx, query, []any{name}, rows); err != nil {
		return false, fmt.Errorf("spanner: checking table existence %w", err)
	}
	defer rows.Close()
	return rows.Next(), nil
}

// maxStringSize defines the maximum size of string types in Spanner.
const maxStringSize = 2621440 // 2.5MB

func (d *Spanner) atOpen(conn dialect.ExecQuerier) (migrate.Driver, error) {
	return nil, fmt.Errorf("spanner: atlas community edition does not support opening spanner connections yet")
}

func (d *Spanner) atTable(t1 *Table, t2 *schema.Table) {
	// Set table name
	t2.Name = t1.Name

	// Spanner doesn't use traditional schemas like PostgreSQL/MySQL
	// Instead, tables exist directly in the database
	if d.schema != "" {
		t2.Schema = schema.New(d.schema)
	}

	// Add Spanner-specific table attributes if needed
	if t1.Annotation != nil {
		// Handle Spanner-specific table options like:
		// - Interleave tables (parent-child relationships)
		// - Row deletion policies
		// - Column families (for performance optimization)
		// These would be implemented here when needed
	}
}

func (d *Spanner) supportsDefault(c *Column) bool {
	// Spanner has limited support for default values
	// Generally supports defaults for most basic types except:
	// - ARRAY types
	// - STRUCT types
	// - JSON with complex expressions
	switch c.Type {
	case field.TypeBool, field.TypeInt8, field.TypeInt16, field.TypeInt32, field.TypeInt, field.TypeInt64,
		field.TypeUint8, field.TypeUint16, field.TypeUint32, field.TypeUint, field.TypeUint64,
		field.TypeFloat32, field.TypeFloat64, field.TypeString, field.TypeBytes:
		return true
	case field.TypeTime:
		// Spanner supports CURRENT_TIMESTAMP() as default for TIMESTAMP
		return true
	case field.TypeJSON:
		// Spanner supports simple JSON defaults
		return true
	case field.TypeEnum, field.TypeUUID:
		// These map to STRING in Spanner, so defaults are supported
		return true
	case field.TypeOther:
		// NUMERIC type supports defaults
		if strings.Contains(strings.ToLower(c.typ), "numeric") || strings.Contains(strings.ToLower(c.typ), "decimal") {
			return true
		}
		return false
	default:
		return false
	}
}

func (d *Spanner) atTypeC(c1 *Column, c2 *schema.Column) error {
	// TODO: Spanner-specific type mapping based on c1.SchemaType
	if c1.SchemaType != nil && c1.SchemaType[dialect.MySQL] != "" {
		return nil
	}

	var t schema.Type
	switch c1.Type {
	case field.TypeBool:
		t = &schema.BoolType{T: "BOOL"}
	case field.TypeInt8, field.TypeInt16, field.TypeInt32, field.TypeInt, field.TypeInt64:
		// Spanner only supports INT64 for integer types
		t = &schema.IntegerType{T: "INT64"}
	case field.TypeUint8, field.TypeUint16, field.TypeUint32, field.TypeUint, field.TypeUint64:
		// Spanner doesn't have unsigned types, map to INT64
		t = &schema.IntegerType{T: "INT64"}
	case field.TypeFloat32:
		// Spanner supports single precision floating point
		t = &schema.FloatType{T: "FLOAT32"}
	case field.TypeFloat64:
		// Spanner supports double precision floating point
		t = &schema.FloatType{T: "FLOAT64"}
	case field.TypeString:
		size := c1.Size
		if size == 0 {
			size = DefaultStringLen
		}
		if size > maxStringSize {
			t = &schema.StringType{T: "STRING(MAX)"}
		} else {
			t = &schema.StringType{T: "STRING", Size: int(size)}
		}
	case field.TypeBytes:
		size := c1.Size
		if size == 0 || size > maxStringSize {
			t = &schema.BinaryType{T: "BYTES(MAX)"}
		} else {
			sizePtr := int(size)
			t = &schema.BinaryType{T: "BYTES", Size: &sizePtr}
		}
	case field.TypeTime:
		// Map to TIMESTAMP - Spanner's absolute point in time type
		// Note: Spanner also has DATE type for calendar dates without time
		t = &schema.TimeType{T: "TIMESTAMP"}
	case field.TypeJSON:
		// Spanner has native JSON support
		t = &schema.JSONType{T: "JSON"}
	case field.TypeEnum:
		// Spanner has ENUM support but requires protocol buffer definition
		// TODO: Implement full ENUM support with proto definitions
		t = &schema.EnumType{T: "ENUM", Values: c1.Enums}
	case field.TypeUUID:
		// Spanner doesn't have native UUID, use STRING(36)
		t = &schema.StringType{T: "STRING", Size: 36}
	default:
		// TODO: Use ParseType as fallback for unknown types
		c2.Type.Type = t
	}
	c2.Type.Type = t
	return nil
}

func (d *Spanner) atUniqueC(t1 *Table, c1 *Column, t2 *schema.Table, c2 *schema.Column) {
	// For UNIQUE columns, Spanner creates an implicit index
	for _, idx := range t1.Indexes {
		// Index also defined explicitly, and will be added in atIndexes.
		if idx.Unique && len(idx.Columns) == 1 && idx.Columns[0].Name == c1.Name {
			return
		}
	}
	t2.AddIndexes(schema.NewUniqueIndex(fmt.Sprintf("%s_unique_%s", t1.Name, c1.Name)).AddColumns(c2))
}

func (d *Spanner) atIncrementC(t *schema.Table, c *schema.Column) {
	// Spanner doesn't support auto-increment like MySQL
	// This is a no-op for Spanner since it doesn't support auto-increment
}

func (d *Spanner) atIncrementT(t *schema.Table, v int64) {
	// Spanner doesn't support table-level auto-increment
	// This is a no-op for Spanner
}

func (d *Spanner) atIndex(idx1 *Index, t2 *schema.Table, idx2 *schema.Index) error {
	for _, c1 := range idx1.Columns {
		c2, ok := t2.Column(c1.Name)
		if !ok {
			return fmt.Errorf("unexpected index %q column: %q", idx1.Name, c1.Name)
		}
		part := &schema.IndexPart{C: c2}
		idx2.AddParts(part)
	}

	// Add Spanner-specific index attributes if needed
	if idx1.Annotation != nil {
		// Handle Spanner-specific index options here:
		// - STORING clause for covering indexes
		// - INTERLEAVE IN PARENT for interleaved indexes
		// - NULL_FILTERED for excluding NULL values
		// For now, this is a placeholder for future enhancements
	}

	// Note: Spanner has restrictions on index key columns:
	// - FLOAT32, ARRAY, JSON, STRUCT are not valid key column types
	// - These restrictions are handled at the DDL level

	return nil
}

func (d *Spanner) atTypeRangeSQL(ts ...string) string {
	for i := range ts {
		ts[i] = fmt.Sprintf("('%s')", ts[i])
	}
	return fmt.Sprintf("INSERT INTO `%s` (`type`) VALUES %s", TypeTable, strings.Join(ts, ", "))
}
