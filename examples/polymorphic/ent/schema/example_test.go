// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"testing"

	"entgo.io/ent/schema/edge"
)

// TestPolymorphicEdgeExample demonstrates the polymorphic edge functionality.
func TestPolymorphicEdgeExample(t *testing.T) {
	t.Run("RegularPolymorphicEdge", func(t *testing.T) {
		// Test regular polymorphic edge (allows multiple instances)
		audioEdge := edge.PolyTo("entity", Media.Type, MediaEpisode.Type, Trailer.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Comment("Regular polymorphic edge - multiple instances allowed")

		desc := audioEdge.Descriptor()
		if !desc.IsPolymorphic {
			t.Error("Expected IsPolymorphic to be true")
		}
		if desc.Unique {
			t.Error("Expected Unique to be false for regular polymorphic edge")
		}
		if len(desc.AllowedTypes) != 3 {
			t.Errorf("Expected 3 polymorphic types, got %d", len(desc.AllowedTypes))
		}
	})

	t.Run("UniquePolymorphicEdge", func(t *testing.T) {
		// Test unique polymorphic edge (only one instance per foreign key + type combination)
		preferenceEdge := edge.PolyTo("entity", Media.Type, MediaEpisode.Type).
			TypeField("entity_type").
			Field("entity_nid").
			Required().
			Unique().
			Comment("Unique polymorphic edge - creates unique constraint")

		desc := preferenceEdge.Descriptor()
		if !desc.IsPolymorphic {
			t.Error("Expected IsPolymorphic to be true")
		}
		if !desc.Unique {
			t.Error("Expected Unique to be true for unique polymorphic edge")
		}
		if len(desc.AllowedTypes) != 2 {
			t.Errorf("Expected 2 polymorphic types, got %d", len(desc.AllowedTypes))
		}
		if desc.TypeDiscriminatorField != "entity_type" {
			t.Errorf("Expected TypeField to be 'entity_type', got %s", desc.TypeDiscriminatorField)
		}
		if desc.Field != "entity_nid" {
			t.Errorf("Expected Field to be 'entity_nid', got %s", desc.Field)
		}
	})
}
