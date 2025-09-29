// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AudioConsumable demonstrates a polymorphic edge.
// It can be linked to Media, MediaEpisode, Trailer, or BonusMaterial.
type AudioConsumable struct {
	ent.Schema
}

func (AudioConsumable) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_nid").Comment("Polymorphic foreign key"),
		field.String("entity_type").Comment("Type discriminator"),
		field.String("format").Comment("Audio format like mp3, aac, etc."),
		field.Int("bitrate").Optional(),
		field.Int("duration_seconds").Optional(),
	}
}

func (AudioConsumable) Edges() []ent.Edge {
	return []ent.Edge{
		// Polymorphic edge that can connect to multiple entity types using .Type
		edge.PolyTo("entity", Media.Type, MediaEpisode.Type, Trailer.Type, BonusMaterial.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Comment("The entity this audio consumable is associated with"),
	}
}
