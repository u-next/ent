package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Annotations of the Product.
func (Product) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Products"},
	}
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.String("name"),
		field.String("description").
			Optional(),
		field.String("sku").
			Unique(),
		field.Float("price"),
		field.String("category").
			Optional(),
		field.Bool("is_active").
			Default(true),
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O inverse relationship: product is manufactured by company
		edge.From("manufacturer", Company.Type).
			Ref("products").
			Unique(),
		// M2M inverse relationship: product is bought by persons through Purchase edge table
		edge.From("buyers", Person.Type).
			Ref("purchases").
			Through("person_purchases", Purchase.Type),
		// M2M self-referential relationship (recommendations)
		edge.To("recommendations", Product.Type).
			From("recommended_by"),
	}
}
