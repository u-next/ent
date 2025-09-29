package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ServiceRecord holds the schema definition for the ServiceRecord entity.
type ServiceRecord struct {
	ent.Schema
}

// Annotations of the ServiceRecord.
func (ServiceRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ServiceRecords"},
	}
}

// Fields of the ServiceRecord.
func (ServiceRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Time("service_date").
			Default(time.Now),
		field.String("service_type").
			Default("maintenance"),
		field.String("description").
			Optional(),
		field.Float("cost").
			Optional(),
		field.Int("mileage").
			Optional(),
		field.String("technician").
			Optional(),
	}
}

// Edges of the ServiceRecord.
func (ServiceRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O inverse relationship: service record belongs to a vehicle
		edge.From("vehicle", Vehicle.Type).
			Ref("service_history").
			Unique().
			Required(),
		// M2O inverse relationship: service record is provided by a service center (company)
		edge.From("service_center", Company.Type).
			Ref("service_records").
			Unique().
			Required(),
	}
}
