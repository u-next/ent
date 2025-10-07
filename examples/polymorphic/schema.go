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

// AudioConsumable demonstrates a polymorphic edge using ToOneOf.
// It can be linked to Media, MediaEpisode, Trailer, or BonusMaterial.
type AudioConsumable struct {
	ent.Schema
}

func (AudioConsumable) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_type").Comment("Type discriminator"),
		field.String("format").Comment("Audio format like mp3, aac, etc."),
		field.Int("bitrate").Optional(),
		field.Int("duration_seconds").Optional(),
	}
}

func (AudioConsumable) Edges() []ent.Edge {
	return []ent.Edge{
		edge.ToOneOf(
			"entity_type",
			"media", Media.Type,
			"media_episode", MediaEpisode.Type,
			"trailer", Trailer.Type,
			"bonus_material", BonusMaterial.Type,
		).Required().Comment("The entity this audio consumable is associated with"),
	}
}

// VideoConsumable demonstrates another polymorphic edge.
type VideoConsumable struct {
	ent.Schema
}

func (VideoConsumable) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_type").Comment("Type discriminator"),
		field.String("format").Comment("Video format like mp4, avi, etc."),
		field.String("resolution").Optional(),
		field.Int("bitrate").Optional(),
		field.Int("duration_seconds").Optional(),
	}
}

func (VideoConsumable) Edges() []ent.Edge {
	return []ent.Edge{
		edge.ToOneOf(
			"entity_type",
			"media", Media.Type,
			"media_episode", MediaEpisode.Type,
			"trailer", Trailer.Type,
			"bonus_material", BonusMaterial.Type,
		).Required().Comment("The entity this video consumable is associated with"),
	}
}

// AvailabilitySummary demonstrates yet another polymorphic edge.
type AvailabilitySummary struct {
	ent.Schema
}

func (AvailabilitySummary) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_type").Comment("Type discriminator"),
		field.String("platform_code").Comment("Platform identifier"),
		field.Bool("available").Default(false),
		field.Time("available_from").Optional(),
		field.Time("available_until").Optional(),
	}
}

func (AvailabilitySummary) Edges() []ent.Edge {
	return []ent.Edge{
		edge.ToOneOf(
			"entity_type",
			"media", Media.Type,
			"media_episode", MediaEpisode.Type,
			"trailer", Trailer.Type,
		).Required().Comment("The entity this availability summary is for"),
	}
}

// Credit demonstrates a polymorphic edge with different target types.
type Credit struct {
	ent.Schema
}

func (Credit) Fields() []ent.Field {
	return []ent.Field{
		field.String("entity_type").Comment("Type discriminator"),
		field.String("role").Comment("Role in the production"),
		field.String("character").Optional().Comment("Character name if applicable"),
		field.Int("order").Optional().Comment("Display order"),
	}
}

func (Credit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.ToOneOf(
			"entity_type",
			"media", Media.Type,
			"media_episode", MediaEpisode.Type,
		).Required().Comment("The entity this credit is associated with"),
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
		field.String("entity_type").Comment("Type discriminator"),
		field.String("preference_value").Comment("The preference setting"),
		field.Time("created_at").Default(func() time.Time { return time.Now() }),
	}
}

func (UserPreference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.ToOneOf(
			"entity_type",
			"media", Media.Type,
			"media_episode", MediaEpisode.Type,
			"trailer", Trailer.Type,
		).Required().Unique().Comment("The entity this preference is for (unique per user per entity)"),
	}
}

// TextMessage represents a text message.
type TextMessage struct {
	ent.Schema
}

func (TextMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("content"),
		field.String("encoding").Default("utf-8"),
	}
}

func (TextMessage) Edges() []ent.Edge {
	return []ent.Edge{}
}

// PhotoMessage represents a photo message.
type PhotoMessage struct {
	ent.Schema
}

func (PhotoMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("photo_url"),
		field.String("caption").Optional(),
		field.Int("width").Optional(),
		field.Int("height").Optional(),
	}
}

func (PhotoMessage) Edges() []ent.Edge {
	return []ent.Edge{}
}

// VideoMessage represents a video message.
type VideoMessage struct {
	ent.Schema
}

func (VideoMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("video_url"),
		field.String("thumbnail_url").Optional(),
		field.Int("duration_seconds").Optional(),
	}
}

func (VideoMessage) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Message demonstrates ToOneOf polymorphic edge.
type Message struct {
	ent.Schema
}

func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.String("sender_id"),
		field.String("message_type"),
		field.Time("created_at").Default(func() time.Time { return time.Now() }),
	}
}

func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.ToOneOf(
			"message_type",
			"text_message", TextMessage.Type,
			"photo_message", PhotoMessage.Type,
			"video_message", VideoMessage.Type,
		).Required().Comment("Polymorphic reference to message content"),
	}
}

// LocalUser represents a local user account.
type LocalUser struct {
	ent.Schema
}

func (LocalUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique(),
		field.String("email").Unique(),
		field.String("password_hash"),
	}
}

func (LocalUser) Edges() []ent.Edge {
	return []ent.Edge{}
}

// ForeignUser represents a user from external system.
type ForeignUser struct {
	ent.Schema
}

func (ForeignUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("external_id").Unique(),
		field.String("provider"),
		field.String("display_name"),
	}
}

func (ForeignUser) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Object demonstrates FromOneOf polymorphic inverse edge.
type Object struct {
	ent.Schema
}

func (Object) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("owner_user_type"),
		field.String("description").Optional(),
	}
}

func (Object) Edges() []ent.Edge {
	return []ent.Edge{
		edge.FromOneOf(
			"owner_user_type",
			"local_user", LocalUser.Type,
			"foreign_user", ForeignUser.Type,
		).Unique().Required().Comment("Polymorphic reference to object owner"),
	}
}

// TaskTemplate demonstrates multiple polymorphic edges.
type TaskTemplate struct {
	ent.Schema
}

func (TaskTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("assignee_type"),
		field.String("reporter_type"),
		field.String("description").Optional(),
	}
}

func (TaskTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.ToOneOf(
			"assignee_type",
			"local_user", LocalUser.Type,
			"foreign_user", ForeignUser.Type,
		).Comment("User assigned to this task"),
		
		edge.ToOneOf(
			"reporter_type", 
			"local_user", LocalUser.Type,
			"foreign_user", ForeignUser.Type,
		).Required().Comment("User who reported this task"),
	}
}
