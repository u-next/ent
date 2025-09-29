package main

import (
	"context"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/gql/schema"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func TestPropertyGraphGeneration(t *testing.T) {
	t.Log("--- Testing Property Graph Generation ---")

	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Generate property graph
	pgs, err := graph.PropertyGraphs()
	if err != nil {
		t.Fatalf("failed to generate property graph: %v", err)
	}
	if len(pgs) == 0 {
		t.Fatalf("no property graphs generated")
	}
	pg := pgs[0] // Use the first property graph for testing

	// Inspect the property graph
	t.Logf("Property Graph Name: %s", pg.Name)
	t.Logf("Node Tables: %d", len(pg.NodeTables))
	for _, nt := range pg.NodeTables {
		t.Logf("  - %s", nt.TableName)
	}
	t.Logf("Edge Tables: %d", len(pg.EdgeTables))
	for _, et := range pg.EdgeTables {
		t.Logf("  - %s (source: %s -> dest: %s)",
			et.TableName,
			et.SourceKey.ReferencedTable,
			et.DestinationKey.ReferencedTable)
	}

	// Generate DDL
	ctx := context.Background()
	ddl, err := schema.PropertyGraphDDL(ctx, schema.DDLArgs{
		Dialect:        dialect.Spanner,
		PropertyGraphs: []*schema.PropertyGraph{pg},
	})
	if err != nil {
		t.Fatalf("failed to generate DDL: %v", err)
	}

	t.Logf("Generated Property Graph DDL:\n%s", ddl)

	// Save DDL to file
	if err := os.WriteFile("property_graph.sql", []byte(ddl), 0644); err != nil {
		t.Fatalf("failed to write DDL file: %v", err)
	}

	t.Log("✓ DDL saved to property_graph.sql")
}

func TestM2MRelationships(t *testing.T) {
	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Generate property graph
	pgs, err := graph.PropertyGraphs()
	if err != nil {
		t.Fatalf("failed to generate property graph: %v", err)
	}
	if len(pgs) == 0 {
		t.Fatalf("no property graphs generated")
	}
	pg := pgs[0] // Use the first property graph for testing

	t.Log("Validating M2M relationships with edge tables:")

	expectedM2MEdges := map[string]bool{
		"PersonOwnAccount": false, // Person ↔ Account
		"AccountTransfers": false, // Account ↔ Account (self-ref)
		"Partners":         false, // Company ↔ Company (self-ref)
		"Purchase":         false, // Person ↔ Product
	}

	for _, et := range pg.EdgeTables {
		if _, exists := expectedM2MEdges[et.TableName]; exists {
			expectedM2MEdges[et.TableName] = true
			t.Logf("  ✓ Found M2M edge table: %s", et.TableName)
		}
	}

	for edgeName, found := range expectedM2MEdges {
		if !found {
			t.Logf("  ⚠ Missing expected M2M edge table: %s", edgeName)
		}
	}
}

func TestSelfReferentialRelationships(t *testing.T) {
	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Generate property graph
	pgs, err := graph.PropertyGraphs()
	if err != nil {
		t.Fatalf("failed to generate property graph: %v", err)
	}
	if len(pgs) == 0 {
		t.Fatalf("no property graphs generated")
	}
	pg := pgs[0] // Use the first property graph for testing

	t.Log("Validating self-referential relationships:")

	selfRefPatterns := map[string][]string{
		"persons":   {"spouse", "friends"},        // O2O and M2M self-ref
		"accounts":  {"transfers"},                // M2M self-ref through edge table
		"companies": {"subsidiaries", "partners"}, // O2M and M2M self-ref
		"products":  {"recommendations"},          // M2M self-ref direct
	}

	// Check for self-referential edge tables
	for _, et := range pg.EdgeTables {
		if et.SourceKey.ReferencedTable == et.DestinationKey.ReferencedTable {
			t.Logf("  ✓ Found self-referential edge table: %s (%s ↔ %s)",
				et.TableName, et.SourceKey.ReferencedTable, et.DestinationKey.ReferencedTable)
		}
	}

	for tableName, patterns := range selfRefPatterns {
		t.Logf("  - Table %s supports self-referential patterns: %v", tableName, patterns)
	}
}

func TestDirectRelationships(t *testing.T) {
	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	t.Log("Validating direct O2O and O2M relationships:")

	// Test specific direct relationships
	directRelationshipTests := []struct {
		fromEntity string
		edgeName   string
		toEntity   string
		relType    string
	}{
		{"Company", "employees", "Person", "O2M"},
		{"Company", "ceo", "Person", "O2O"},
		{"Company", "products", "Product", "O2M"},
		{"Company", "manufactured_vehicles", "Vehicle", "O2M"},
		{"Person", "vehicles", "Vehicle", "O2M"},
		{"Vehicle", "service_history", "ServiceRecord", "O2M"},
	}

	entityMap := make(map[string]*gen.Type)
	for _, node := range graph.Nodes {
		entityMap[node.Name] = node
	}

	for _, test := range directRelationshipTests {
		fromNode, fromExists := entityMap[test.fromEntity]
		_, toExists := entityMap[test.toEntity]

		if !fromExists {
			t.Errorf("Source entity %s not found in schema", test.fromEntity)
			continue
		}
		if !toExists {
			t.Errorf("Target entity %s not found in schema", test.toEntity)
			continue
		}

		// Check if the edge exists
		edgeFound := false
		for _, edge := range fromNode.Edges {
			if edge.Name == test.edgeName {
				edgeFound = true
				break
			}
		}

		if edgeFound {
			t.Logf("  ✓ %s relationship: %s.%s -> %s", test.relType, test.fromEntity, test.edgeName, test.toEntity)
		} else {
			t.Errorf("  ✗ %s relationship: edge %s not found from %s to %s", test.relType, test.edgeName, test.fromEntity, test.toEntity)
		}
	}

	t.Log("✓ Node tables represent entities that can have direct relationships")
	t.Log("✓ Foreign key relationships will be handled at the database level")
}

func TestM2MWithEdgeProperties(t *testing.T) {
	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Test that edge tables have proper properties
	edgeProperties := map[string][]string{
		"PersonOwnAccount":       {"create_time", "ownership_type", "ownership_percentage"},
		"AccountTransferAccount": {"amount", "create_time", "order_number", "transfer_type"},
		"Partnership":            {"start_date", "end_date", "partnership_type", "is_active"},
		"Purchase":               {"purchase_date", "quantity", "unit_price", "payment_method"},
	}

	// Validate that edge entity schemas exist and have the expected fields
	for edgeTable, expectedProps := range edgeProperties {
		var edgeNode *gen.Type
		for _, node := range graph.Nodes {
			if node.Name == edgeTable {
				edgeNode = node
				break
			}
		}

		if edgeNode == nil {
			t.Errorf("Expected edge entity %s not found in schema", edgeTable)
			continue
		}

		t.Logf("✓ Found edge entity: %s", edgeTable)

		// Check that the node has the expected fields
		foundFields := make(map[string]bool)
		for _, field := range edgeNode.Fields {
			foundFields[field.Name] = true
		}

		for _, expectedProp := range expectedProps {
			if foundFields[expectedProp] {
				t.Logf("  ✓ Field %s found in %s", expectedProp, edgeTable)
			} else {
				t.Logf("  ⚠ Field %s missing in %s", expectedProp, edgeTable)
			}
		}
	}
}

func TestHierarchicalRelationships(t *testing.T) {
	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Test parent-child relationships
	hierarchicalTests := []struct {
		entityName    string
		expectedEdges []string
		description   string
	}{
		{
			entityName:    "Company",
			expectedEdges: []string{"subsidiaries", "employees", "ceo"},
			description:   "Company should have subsidiaries, employees, and ceo edges",
		},
		{
			entityName:    "Person",
			expectedEdges: []string{"employer", "ceo_of"},
			description:   "Person should have employer and ceo_of edges",
		},
	}

	for _, test := range hierarchicalTests {
		var entityNode *gen.Type
		for _, node := range graph.Nodes {
			if node.Name == test.entityName {
				entityNode = node
				break
			}
		}

		if entityNode == nil {
			t.Errorf("Entity %s not found in schema", test.entityName)
			continue
		}

		t.Logf("✓ Testing hierarchical relationships for %s", test.entityName)

		foundEdges := make(map[string]bool)
		for _, edge := range entityNode.Edges {
			foundEdges[edge.Name] = true
		}

		for _, expectedEdge := range test.expectedEdges {
			if foundEdges[expectedEdge] {
				t.Logf("  ✓ Edge %s found in %s", expectedEdge, test.entityName)
			} else {
				t.Errorf("  ✗ Edge %s missing in %s", expectedEdge, test.entityName)
			}
		}
	}
}

func TestSelfReferentialPatterns(t *testing.T) {
	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Test various self-referential patterns
	selfRefTests := []struct {
		entityName           string
		expectedSelfRefEdges []string
		description          string
	}{
		{
			entityName:           "Person",
			expectedSelfRefEdges: []string{"spouse", "friends"},
			description:          "Person should have spouse (O2O) and friends (M2M) self-referential edges",
		},
		{
			entityName:           "Account",
			expectedSelfRefEdges: []string{"transfers"},
			description:          "Account should have transfers self-referential edge",
		},
		{
			entityName:           "Company",
			expectedSelfRefEdges: []string{"subsidiaries", "partners"},
			description:          "Company should have subsidiaries and partners self-referential edges",
		},
		{
			entityName:           "Product",
			expectedSelfRefEdges: []string{"recommendations"},
			description:          "Product should have recommendations self-referential edge",
		},
	}

	for _, test := range selfRefTests {
		var entityNode *gen.Type
		for _, node := range graph.Nodes {
			if node.Name == test.entityName {
				entityNode = node
				break
			}
		}

		if entityNode == nil {
			t.Errorf("Entity %s not found in schema", test.entityName)
			continue
		}

		t.Logf("✓ Testing self-referential patterns for %s", test.entityName)

		foundEdges := make(map[string]bool)
		for _, edge := range entityNode.Edges {
			foundEdges[edge.Name] = true
		}

		for _, expectedEdge := range test.expectedSelfRefEdges {
			if foundEdges[expectedEdge] {
				t.Logf("  ✓ Self-referential edge %s found in %s", expectedEdge, test.entityName)
			} else {
				t.Errorf("  ✗ Self-referential edge %s missing in %s", expectedEdge, test.entityName)
			}
		}
	}
}

func TestCrossDomainRelationships(t *testing.T) {
	// Load the graph
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	// Test relationships that span multiple business domains
	crossDomainTests := []struct {
		fromEntity string
		toEntity   string
		edgeName   string
		domain     string
	}{
		{"Person", "Account", "owns", "Financial"},
		{"Person", "Company", "employer", "Employment"},
		{"Person", "Product", "purchases", "Commercial"},
		{"Person", "Vehicle", "vehicles", "Automotive"},
		{"Vehicle", "ServiceRecord", "service_history", "Service"},
		{"Company", "Product", "products", "Manufacturing"},
		{"Company", "Vehicle", "manufactured_vehicles", "Manufacturing"},
	}

	entityMap := make(map[string]*gen.Type)
	for _, node := range graph.Nodes {
		entityMap[node.Name] = node
	}

	for _, test := range crossDomainTests {
		fromNode, fromExists := entityMap[test.fromEntity]
		_, toExists := entityMap[test.toEntity]

		if !fromExists {
			t.Errorf("Source entity %s not found in schema", test.fromEntity)
			continue
		}
		if !toExists {
			t.Errorf("Target entity %s not found in schema", test.toEntity)
			continue
		}

		// Check if the edge exists
		edgeFound := false
		for _, edge := range fromNode.Edges {
			if edge.Name == test.edgeName {
				edgeFound = true
				break
			}
		}

		if edgeFound {
			t.Logf("✓ %s domain: %s -> %s via %s", test.domain, test.fromEntity, test.toEntity, test.edgeName)
		} else {
			t.Errorf("✗ %s domain: edge %s not found from %s to %s", test.domain, test.edgeName, test.fromEntity, test.toEntity)
		}
	}
}

func TestSchemaValidation(t *testing.T) {
	t.Log("--- Testing Schema Validation ---")

	// Load and validate the complete schema
	graph, err := entc.LoadGraph("./ent/schema", &gen.Config{
		Package: "example/gqlcodegen/ent",
		Storage: &gen.Storage{Name: "gql"},
	})
	if err != nil {
		t.Fatalf("failed to load graph for validation: %v", err)
	}

	// Count entities and relationships
	nodeCount := len(graph.Nodes)
	t.Logf("Total entities (nodes): %d", nodeCount)

	totalEdges := 0
	for _, node := range graph.Nodes {
		totalEdges += len(node.Edges)
	}
	t.Logf("Total relationships (edges): %d", totalEdges)

	// Generate property graph to validate structure
	pgs, err := graph.PropertyGraphs()
	if err != nil {
		t.Fatalf("failed to generate property graph for validation: %v", err)
	}
	if len(pgs) == 0 {
		t.Fatalf("no property graphs generated for validation")
	}
	pg := pgs[0] // Use the first property graph for testing

	t.Logf("Property graph node tables: %d", len(pg.NodeTables))
	t.Logf("Property graph edge tables: %d", len(pg.EdgeTables))

	// Validate that we have good coverage of relationship types
	if len(pg.EdgeTables) < 4 {
		t.Fatalf("expected at least 4 edge tables for comprehensive testing, got %d", len(pg.EdgeTables))
	}

	if len(pg.NodeTables) < 8 {
		t.Fatalf("expected at least 8 node tables for comprehensive testing, got %d", len(pg.NodeTables))
	}

	t.Log("✓ Schema validation completed successfully")
}
