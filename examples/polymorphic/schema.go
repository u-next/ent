// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package polymorphic

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Media represents a media entity.
type Media struct {
	ent.Schema
}

func (Media) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.String("description").Optional(),
	}
}

func (Media) Edges() []ent.Edge {
	return []ent.Edge{}
}

// MediaEpisode represents an episode of a media.
type MediaEpisode struct {
	ent.Schema
}

func (MediaEpisode) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Int("episode_number"),
		field.String("description").Optional(),
	}
}

func (MediaEpisode) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Trailer represents a trailer for media content.
type Trailer struct {
	ent.Schema
}

func (Trailer) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.String("url"),
		field.Int("duration_seconds"),
	}
}

func (Trailer) Edges() []ent.Edge {
	return []ent.Edge{}
}

// BonusMaterial represents bonus content.
type BonusMaterial struct {
	ent.Schema
}

func (BonusMaterial) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.String("description").Optional(),
	}
}

func (BonusMaterial) Edges() []ent.Edge {
	return []ent.Edge{}
}

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
		// Polymorphic edge that can connect to multiple entity types
		edge.PolyTo("entity", Media.Type, MediaEpisode.Type, Trailer.Type, BonusMaterial.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Comment("The entity this audio consumable is associated with"),
	}
}

// VideoConsumable demonstrates another polymorphic edge.
type VideoConsumable struct {
	ent.Schema
}

func (VideoConsumable) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_nid").Comment("Polymorphic foreign key"),
		field.String("entity_type").Comment("Type discriminator"),
		field.String("format").Comment("Video format like mp4, avi, etc."),
		field.String("resolution").Optional(),
		field.Int("bitrate").Optional(),
		field.Int("duration_seconds").Optional(),
	}
}

func (VideoConsumable) Edges() []ent.Edge {
	return []ent.Edge{
		// Another polymorphic edge with same target types
		edge.PolyTo("entity", Media.Type, MediaEpisode.Type, Trailer.Type, BonusMaterial.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Comment("The entity this video consumable is associated with"),
	}
}

// AvailabilitySummary demonstrates yet another polymorphic edge.
type AvailabilitySummary struct {
	ent.Schema
}

func (AvailabilitySummary) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_nid").Comment("Polymorphic foreign key"),
		field.String("entity_type").Comment("Type discriminator"),
		field.String("platform_code").Comment("Platform identifier"),
		field.Bool("available").Default(false),
		field.Time("available_from").Optional(),
		field.Time("available_until").Optional(),
	}
}

func (AvailabilitySummary) Edges() []ent.Edge {
	return []ent.Edge{
		// Polymorphic edge to availability entities
		edge.PolyTo("entity", Media.Type, MediaEpisode.Type, Trailer.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Comment("The entity this availability summary is for"),
	}
}

// Credit demonstrates a polymorphic edge with different target types.
type Credit struct {
	ent.Schema
}

func (Credit) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_nid").Comment("Polymorphic foreign key"),
		field.String("entity_type").Comment("Type discriminator"),
		field.String("role").Comment("Role in the production"),
		field.String("character").Optional().Comment("Character name if applicable"),
		field.Int("order").Optional().Comment("Display order"),
	}
}

func (Credit) Edges() []ent.Edge {
	return []ent.Edge{
		// Polymorphic edge for credits
		edge.PolyTo("entity", Media.Type, MediaEpisode.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Comment("The entity this credit is associated with"),
	}
}

// UserPreference demonstrates a unique polymorphic edge.
// Each user can have only one preference per entity type.
type UserPreference struct {
	ent.Schema
}

func (UserPreference) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Comment("User identifier"),
		field.String("entity_nid").Comment("Polymorphic foreign key"),
		field.String("entity_type").Comment("Type discriminator"),
		field.String("preference_value").Comment("The preference setting"),
		field.Time("created_at").Default(func() time.Time { return time.Now() }),
	}
}

func (UserPreference) Edges() []ent.Edge {
	return []ent.Edge{
		// Unique polymorphic edge - each user can have only one preference per entity
		edge.PolyTo("entity", Media.Type, MediaEpisode.Type, Trailer.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Unique().
			Comment("The entity this preference is for (unique per user per entity)"),
	}
}
