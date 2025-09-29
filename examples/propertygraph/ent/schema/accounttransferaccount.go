package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AccountTransferAccount holds the schema definition for the AccountTransferAccount edge table.
type AccountTransferAccount struct {
	ent.Schema
}

// Annotations of the AccountTransferAccount.
func (AccountTransferAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "AccountTransferAccounts"},
	}
}

// Fields of the AccountTransferAccount.
func (AccountTransferAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.Int64("from_account_id"),
		field.Int64("to_account_id"),
		field.Float("amount"),
		field.Time("create_time").
			Default(time.Now).
			Immutable(),
		field.String("order_number").
			Unique(),
		field.String("transfer_type").
			Default("wire"),
		field.String("memo").
			Optional(),
	}
}

// Edges of the AccountTransferAccount.
func (AccountTransferAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("from_account", Account.Type).
			Unique().
			Required().
			Field("from_account_id"),
		edge.To("to_account", Account.Type).
			Unique().
			Required().
			Field("to_account_id"),
	}
}
