package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PersonOwnAccount holds the schema definition for the PersonOwnAccount edge table.
type PersonOwnAccount struct {
	ent.Schema
}

// Annotations of the PersonOwnAccount.
func (PersonOwnAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "PersonOwnAccounts"},
	}
}

// Fields of the PersonOwnAccount.
func (PersonOwnAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.Int64("person_id"),
		field.Int64("account_id"),
		field.Time("create_time").
			Default(time.Now).
			Immutable(),
		field.String("ownership_type").
			Default("owner"),
		field.Float("ownership_percentage").
			Default(100.0),
	}
}

// Edges of the PersonOwnAccount.
func (PersonOwnAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("person", Person.Type).
			Unique().
			Required().
			Field("person_id"),
		edge.To("account", Account.Type).
			Unique().
			Required().
			Field("account_id"),
	}
}
