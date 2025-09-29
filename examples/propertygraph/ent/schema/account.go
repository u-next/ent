package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Account holds the schema definition for the Account entity.
type Account struct {
	ent.Schema
}

// Annotations of the Account.
func (Account) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Accounts"},
	}
}

// Fields of the Account.
func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.Time("create_time").
			Default(time.Now).
			Immutable(),
		field.Bool("is_blocked").
			Default(false),
		field.String("nick_name").
			Optional(),
		field.Float("balance").
			Default(0.0),
		field.String("account_type").
			Default("checking"),
	}
}

// Edges of the Account.
func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		// M2M relationship to Person through PersonOwnAccount edge table
		edge.From("owners", Person.Type).
			Ref("owns").
			Through("person_accounts", PersonOwnAccount.Type),
		// Self-referential edge for transfers (without complex edge table for now)
		edge.To("transfers", Account.Type).StorageKey(edge.Table("AccountTransfers")),
	}
}
