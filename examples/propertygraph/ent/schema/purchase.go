package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Purchase holds the schema definition for the Purchase edge table.
type Purchase struct {
	ent.Schema
}

// Annotations of the Purchase.
func (Purchase) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Purchases"},
	}
}

// Fields of the Purchase.
func (Purchase) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("person_id"),
		field.Int64("product_id"),
		field.Time("purchase_date").
			Default(time.Now),
		field.Int("quantity").
			Default(1),
		field.Float("unit_price"),
		field.Float("total_amount"),
		field.String("payment_method").
			Default("credit_card"),
		field.String("order_status").
			Default("completed"),
	}
}

// Edges of the Purchase.
func (Purchase) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("buyer", Person.Type).
			Unique().
			Required().
			Field("person_id"),
		edge.To("product", Product.Type).
			Unique().
			Required().
			Field("product_id"),
	}
}
