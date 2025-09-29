package printer

import (
	"context"
	"strings"
	"testing"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"

	"github.com/stretchr/testify/assert"
)

func TestGraphviz_Graphviz(t *testing.T) {
	tests := []struct {
		input *gen.Graph
		out   string
	}{
		{
			input: &gen.Graph{
				Nodes: []*gen.Type{
					{
						Name: "User",
						ID:   &gen.Field{Name: "id", Type: &field.TypeInfo{Type: field.TypeInt}},
						Fields: []*gen.Field{
							{Name: "name", Type: &field.TypeInfo{Type: field.TypeString}, Validators: 1},
							{Name: "age", Type: &field.TypeInfo{Type: field.TypeInt}, Nillable: true},
							{Name: "created_at", Type: &field.TypeInfo{Type: field.TypeTime}, Nillable: true, Immutable: true},
						},
					},
				},
			},
			out: `
digraph "" {
	graph [bb="0,0,58.541,36"];
	node [label="\N"];
	User	[height=0.5,
		pos="29.271,18",
		width=0.81307];
}
`,
		},
		{
			input: &gen.Graph{
				Nodes: []*gen.Type{
					{
						Name: "User",
						ID:   &gen.Field{Name: "id", Type: &field.TypeInfo{Type: field.TypeInt}},
						Edges: []*gen.Edge{
							{Name: "groups", Type: &gen.Type{Name: "Group"}, Rel: gen.Relation{Type: gen.M2M}, Optional: true},
							{Name: "spouse", Type: &gen.Type{Name: "User"}, Unique: true, Rel: gen.Relation{Type: gen.O2O}},
						},
					},
				},
			},
			out: `
digraph "" {
	graph [bb="0,0,76.541,36"];
	node [label="\N"];
	User	[height=0.5,
		pos="29.271,18",
		width=0.81307];
	User -> User	[key=spouse,
		pos="e,56.947,11.276 56.947,24.724 67.647,25.022 76.541,22.781 76.541,18 76.541,15.086 73.239,13.116 68.237,12.089"];
}
`,
		},
		{
			input: &gen.Graph{
				Nodes: []*gen.Type{
					{
						Name: "User",
						ID:   &gen.Field{Name: "id", Type: &field.TypeInfo{Type: field.TypeInt}},
						Fields: []*gen.Field{
							{Name: "name", Type: &field.TypeInfo{Type: field.TypeString}, Validators: 1},
							{Name: "age", Type: &field.TypeInfo{Type: field.TypeInt}, Nillable: true},
						},
						Edges: []*gen.Edge{
							{Name: "groups", Type: &gen.Type{Name: "Group"}, Rel: gen.Relation{Type: gen.M2M}, Optional: true},
							{Name: "spouse", Type: &gen.Type{Name: "User"}, Unique: true, Rel: gen.Relation{Type: gen.O2O}},
						},
					},
					{
						Name: "Group",
						ID:   &gen.Field{Name: "id", Type: &field.TypeInfo{Type: field.TypeInt}},
						Fields: []*gen.Field{
							{Name: "name", Type: &field.TypeInfo{Type: field.TypeString}},
						},
						Edges: []*gen.Edge{
							{Name: "users", Type: &gen.Type{Name: "User"}, Rel: gen.Relation{Type: gen.M2M}, Optional: true},
						},
					},
				},
			},
			out: `
digraph "" {
	graph [bb="0,0,82.982,108"];
	node [label="\N"];
	User	[height=0.5,
		pos="35.712,90",
		width=0.81307];
	User -> User	[key=spouse,
		pos="e,57.102,77.453 57.102,102.55 70.363,105.57 82.982,101.39 82.982,90 82.982,81.991 76.743,77.546 68.429,76.664"];
	Group	[height=0.5,
		pos="35.712,18",
		width=0.99199];
	User -> Group	[key=groups,
		pos="e,29.85,35.789 29.833,72.055 29.04,64.574 28.788,55.579 29.078,47.137"];
	Group -> User	[key=users,
		pos="e,41.59,72.055 41.573,35.789 42.375,43.248 42.634,52.237 42.351,60.686"];
}
`,
		},
	}
	for _, tt := range tests {
		b := &strings.Builder{}
		Graphviz(context.Background(), b, tt.input)
		assert.Equal(t, tt.out, "\n"+b.String())
	}
}
