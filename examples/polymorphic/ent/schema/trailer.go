// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

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
