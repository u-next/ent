package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Partnership holds the schema definition for the Partnership edge table.
type Partnership struct {
	ent.Schema
}

// Annotations of the Partnership.
func (Partnership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Partnerships"},
	}
}

// Fields of the Partnership.
func (Partnership) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("company_a_id"),
		field.Int64("company_b_id"),
		field.Time("start_date").
			Default(time.Now),
		field.Time("end_date").
			Optional(),
		field.String("partnership_type").
			Default("strategic"),
		field.String("description").
			Optional(),
		field.Bool("is_active").
			Default(true),
	}
}

// Edges of the Partnership.
func (Partnership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("company_a", Company.Type).
			Unique().
			Required().
			Field("company_a_id"),
		edge.To("company_b", Company.Type).
			Unique().
			Required().
			Field("company_b_id"),
	}
}
