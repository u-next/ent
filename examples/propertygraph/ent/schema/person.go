package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Person holds the schema definition for the Person entity.
type Person struct {
	ent.Schema
}

// Annotations of the Person.
func (Person) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Persons"},
	}
}

// Fields of the Person.
func (Person) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.String("name"),
		field.Time("birthday").
			Optional(),
		field.String("country").
			Optional(),
		field.String("city").
			Optional(),
		field.String("email").
			Unique().
			Optional(),
		field.String("phone").
			Optional(),
	}
}

// Edges of the Person.
func (Person) Edges() []ent.Edge {
	return []ent.Edge{
		// M2M relationship to Account through PersonOwnAccount edge table
		edge.To("owns", Account.Type).
			Through("person_accounts", PersonOwnAccount.Type),
		// O2O self-referential relationship (spouse)
		edge.To("spouse", Person.Type).
			Unique().
			From("spouse"),
		// M2M self-referential relationship (friends)
		edge.To("friends", Person.Type).
			StorageKey(edge.Table("Friends")),
		// M2O inverse relationship: person works for company
		edge.From("employer", Company.Type).
			Ref("employees").
			Unique(),
		// O2O inverse relationship: person is CEO of company
		edge.From("ceo_of", Company.Type).
			Ref("ceo").
			Unique(),
		// M2M relationship to Product through Purchase edge table
		edge.To("purchases", Product.Type).
			Through("person_purchases", Purchase.Type),
		// O2M relationship: Person owns vehicles
		edge.To("vehicles", Vehicle.Type),
	}
}
