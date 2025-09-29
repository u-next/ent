// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

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
