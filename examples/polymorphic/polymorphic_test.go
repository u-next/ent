package polymorphic

import (
	"context"
	"testing"

	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema/field"
)

func TestPolymorphicEdgeGeneration(t *testing.T) {
	// Define base entity schemas
	mediaSchema := &load.Schema{
		Name: "Media",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "title", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "description", Info: &field.TypeInfo{Type: field.TypeString}, Optional: true},
		},
	}

	mediaEpisodeSchema := &load.Schema{
		Name: "MediaEpisode",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "title", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "episode_number", Info: &field.TypeInfo{Type: field.TypeInt}},
		},
	}

	trailerSchema := &load.Schema{
		Name: "Trailer",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "title", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "url", Info: &field.TypeInfo{Type: field.TypeString}},
		},
	}

	bonusMaterialSchema := &load.Schema{
		Name: "BonusMaterial",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "title", Info: &field.TypeInfo{Type: field.TypeString}},
		},
	}

	// Define polymorphic edge schema
	audioConsumableSchema := &load.Schema{
		Name: "AudioConsumable",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "entity_nid", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "entity_type", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "format", Info: &field.TypeInfo{Type: field.TypeString}},
		},
		Edges: []*load.Edge{
			{
				Name:                   "entity",
				Field:                  "entity_nid",
				IsPolymorphic:          true,
				AllowedTypes:           []string{"Media", "MediaEpisode", "Trailer", "BonusMaterial"},
				TypeDiscriminatorField: "entity_type",
				Required:               true,
			},
		},
	}

	// Create the graph
	graph, err := gen.NewGraph(&gen.Config{
		Package: "test/polymorphic",
		Storage: &gen.Storage{Name: "sql"},
	}, mediaSchema, mediaEpisodeSchema, trailerSchema, bonusMaterialSchema, audioConsumableSchema)

	if err != nil {
		t.Fatalf("Failed creating graph: %v", err)
	}

	// Validate the polymorphic edge was processed correctly
	var audioConsumableType *gen.Type
	for _, n := range graph.Nodes {
		if n.Name == "AudioConsumable" {
			audioConsumableType = n
			break
		}
	}

	if audioConsumableType == nil {
		t.Fatal("AudioConsumable type not found")
	}

	// Check that the polymorphic edge was created
	var polymorphicEdge *gen.Edge
	for _, e := range audioConsumableType.Edges {
		if e.Name == "entity" {
			polymorphicEdge = e
			break
		}
	}

	if polymorphicEdge == nil {
		t.Fatal("Polymorphic edge 'entity' not found")
	}

	// Validate polymorphic edge properties
	if !polymorphicEdge.IsPolymorphic {
		t.Error("Edge should be marked as polymorphic")
	}

	if polymorphicEdge.TypeDiscriminatorField != "entity_type" {
		t.Errorf("Expected TypeDiscriminatorField to be 'entity_type', got '%s'", polymorphicEdge.TypeDiscriminatorField)
	}

	expectedTypes := []string{"Media", "MediaEpisode", "Trailer", "BonusMaterial"}
	if len(polymorphicEdge.AllowedTypes) != len(expectedTypes) {
		t.Errorf("Expected %d polymorphic types, got %d", len(expectedTypes), len(polymorphicEdge.AllowedTypes))
	}

	for i, expectedType := range expectedTypes {
		if i >= len(polymorphicEdge.AllowedTypes) {
			t.Errorf("Missing polymorphic type: %s", expectedType)
			continue
		}
		if polymorphicEdge.AllowedTypes[i].Name != expectedType {
			t.Errorf("Expected polymorphic type '%s', got '%s'", expectedType, polymorphicEdge.AllowedTypes[i].Name)
		}
	}

	// Validate relation properties
	if polymorphicEdge.Rel.Type != gen.Polymorphic {
		t.Errorf("Expected relation type Polymorphic, got %v", polymorphicEdge.Rel.Type)
	}

	if polymorphicEdge.Rel.TypeColumn != "entity_type" {
		t.Errorf("Expected TypeColumn to be 'entity_type', got '%s'", polymorphicEdge.Rel.TypeColumn)
	}

	if len(polymorphicEdge.Rel.Columns) != 1 || polymorphicEdge.Rel.Columns[0] != "entity_nid" {
		t.Errorf("Expected Columns to be ['entity_nid'], got %v", polymorphicEdge.Rel.Columns)
	}

	t.Logf("✓ Polymorphic edge validation passed")
}

func TestPolymorphicEdgePropertyGraph(t *testing.T) {
	ctx := context.Background()

	// Simplified schemas for property graph testing
	mediaSchema := &load.Schema{
		Name: "Media",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "title", Info: &field.TypeInfo{Type: field.TypeString}},
		},
	}

	trailerSchema := &load.Schema{
		Name: "Trailer",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "title", Info: &field.TypeInfo{Type: field.TypeString}},
		},
	}

	audioConsumableSchema := &load.Schema{
		Name: "AudioConsumable",
		Fields: []*load.Field{
			{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "entity_nid", Info: &field.TypeInfo{Type: field.TypeString}},
			{Name: "entity_type", Info: &field.TypeInfo{Type: field.TypeString}},
		},
		Edges: []*load.Edge{
			{
				Name:                   "entity",
				Field:                  "entity_nid",
				IsPolymorphic:          true,
				AllowedTypes:           []string{"Media", "Trailer"},
				TypeDiscriminatorField: "entity_type",
				Required:               true,
			},
		},
	}

	// Create the graph
	graph, err := gen.NewGraph(&gen.Config{
		Package: "test/polymorphic",
		Storage: &gen.Storage{Name: "sql"},
	}, mediaSchema, trailerSchema, audioConsumableSchema)

	if err != nil {
		t.Fatalf("Failed creating graph: %v", err)
	}

	// Generate property graph
	pgs, err := graph.PropertyGraphs()
	if err != nil {
		t.Fatalf("Failed generating property graph: %v", err)
	}
	pg := pgs[0]

	ts, err := graph.Tables()
	if err != nil {
		t.Fatalf("Failed generating tables: %v", err)
	}

	// Generate DDL
	ddl, err := schema.DDL(ctx, schema.DDLArgs{
		Dialect:        "spanner",
		PropertyGraphs: pgs,
		Tables:         ts,
	})
	if err != nil {
		t.Fatalf("Failed generating DDL: %v", err)
	}

	t.Logf("Generated Property Graph DDL:\n%s", ddl)

	// Validate that node tables are present
	if len(pg.NodeTables) < 3 {
		t.Errorf("Expected at least 3 node tables, got %d", len(pg.NodeTables))
	}

	nodeTableNames := make(map[string]bool)
	for _, nt := range pg.NodeTables {
		nodeTableNames[nt.TableName] = true
	}

	expectedNodeTables := []string{"Media", "Trailer", "AudioConsumable"}
	for _, expected := range expectedNodeTables {
		if !nodeTableNames[expected] {
			t.Errorf("Missing expected node table: %s", expected)
		}
	}

	t.Logf("✓ Property graph generation passed")
	t.Logf("Node tables: %v", func() []string {
		var names []string
		for _, nt := range pg.NodeTables {
			names = append(names, nt.TableName)
		}
		return names
	}())
}

func TestPolymorphicEdgeValidation(t *testing.T) {
	tests := []struct {
		name        string
		edge        *load.Edge
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid polymorphic edge",
			edge: &load.Edge{
				Name:                   "entity",
				Field:                  "entity_nid",
				IsPolymorphic:          true,
				AllowedTypes:           []string{"Media", "Trailer"},
				TypeDiscriminatorField: "entity_type",
				Required:               true,
			},
			expectError: false,
		},
		{
			name: "missing type field",
			edge: &load.Edge{
				Name:          "entity",
				Field:         "entity_nid",
				IsPolymorphic: true,
				AllowedTypes:  []string{"Media", "Trailer"},
				Required:      true,
			},
			expectError: true,
			errorMsg:    "is missing TypeField",
		},
		{
			name: "missing field",
			edge: &load.Edge{
				Name:                   "entity",
				IsPolymorphic:          true,
				AllowedTypes:           []string{"Media", "Trailer"},
				TypeDiscriminatorField: "entity_type",
				Required:               true,
			},
			expectError: true,
			errorMsg:    "is missing Field",
		},
		{
			name: "no polymorphic types",
			edge: &load.Edge{
				Name:                   "entity",
				Field:                  "entity_nid",
				IsPolymorphic:          true,
				TypeDiscriminatorField: "entity_type",
				Required:               true,
			},
			expectError: true,
			errorMsg:    "must specify at least one target type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple schema with the test edge
			testSchema := &load.Schema{
				Name: "TestEntity",
				Fields: []*load.Field{
					{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}},
					{Name: "entity_nid", Info: &field.TypeInfo{Type: field.TypeString}},
					{Name: "entity_type", Info: &field.TypeInfo{Type: field.TypeString}},
				},
				Edges: []*load.Edge{tt.edge},
			}

			mediaSchema := &load.Schema{
				Name:   "Media",
				Fields: []*load.Field{{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}}},
			}

			trailerSchema := &load.Schema{
				Name:   "Trailer",
				Fields: []*load.Field{{Name: "id", Info: &field.TypeInfo{Type: field.TypeString}}},
			}

			_, err := gen.NewGraph(&gen.Config{
				Package: "test/validation",
				Storage: &gen.Storage{Name: "sql"},
			}, mediaSchema, trailerSchema, testSchema)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test case '%s', but got none", tt.name)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for test case '%s': %v", tt.name, err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
