package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Company holds the schema definition for the Company entity.
type Company struct {
	ent.Schema
}

// Annotations of the Company.
func (Company) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Companies"},
	}
}

// Fields of the Company.
func (Company) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.String("name"),
		field.String("industry").
			Optional(),
		field.String("country").
			Optional(),
		field.Int("employee_count").
			Optional(),
		field.Float("revenue").
			Optional(),
	}
}

// Edges of the Company.
func (Company) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M relationship: Company has many employees
		edge.To("employees", Person.Type),
		// O2O relationship: Company has one CEO
		edge.To("ceo", Person.Type).
			Unique(),
		// Self-referential edge for partners (without complex edge table for now)
		edge.To("partners", Company.Type).
			StorageKey(edge.Table("Partners")),
		// Self-referential O2M: parent company / subsidiaries
		edge.To("subsidiaries", Company.Type).
			From("parent"),
		// O2M relationship: Company manufactures products
		edge.To("products", Product.Type),
		// O2M relationship: Company manufactures vehicles
		edge.To("manufactured_vehicles", Vehicle.Type),
		// O2M relationship: Company provides service records
		edge.To("service_records", ServiceRecord.Type),
	}
}
