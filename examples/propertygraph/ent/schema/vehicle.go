package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Vehicle holds the schema definition for the Vehicle entity.
type Vehicle struct {
	ent.Schema
}

// Annotations of the Vehicle.
func (Vehicle) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Vehicles"},
	}
}

// Fields of the Vehicle.
func (Vehicle) Fields() []ent.Field {
	return []ent.Field{
		field.String("vin").
			Unique(),
		field.String("make"),
		field.String("model"),
		field.Int("year"),
		field.String("color").
			Optional(),
		field.String("vehicle_type").
			Default("car"), // car, truck, motorcycle, etc.
		field.Float("price").
			Optional(),
		field.Bool("is_electric").
			Default(false),
	}
}

// Edges of the Vehicle.
func (Vehicle) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O inverse relationship: vehicle is owned by person
		edge.From("owner", Person.Type).
			Ref("vehicles").
			Unique(),
		// M2O inverse relationship: vehicle is manufactured by company
		edge.From("manufacturer", Company.Type).
			Ref("manufactured_vehicles").
			Unique(),
		// O2M relationship: Vehicle has many service records
		edge.To("service_history", ServiceRecord.Type),
	}
}
