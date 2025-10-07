// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package edge

import (
	"reflect"
	"strings"

	"entgo.io/ent/schema"
)

// A Descriptor for edge configuration.
type Descriptor struct {
	Tag         string                 // struct tag.
	Type        string                 // edge type.
	Name        string                 // edge name.
	Field       string                 // edge field name (e.g. foreign-key).
	RefName     string                 // ref name; inverse only.
	Ref         *Descriptor            // edge reference; to/from of the same type.
	Through     *struct{ N, T string } // through type and name.
	Unique      bool                   // unique edge.
	Inverse     bool                   // inverse edge.
	Required    bool                   // required on creation.
	Immutable   bool                   // create only edge.
	StorageKey  *StorageKey            // optional storage-key configuration.
	Annotations []schema.Annotation    // edge annotations.
	Comment     string                 // edge comment.

	// Polymorphic edge support
	AllowedTypes           []string // types this polymorphic edge can reference.
	TypeDiscriminatorField string   // field name for the polymorphic type discriminator.
	IsPolymorphic          bool     // whether this is a polymorphic edge.
}

// To defines an association edge between two vertices.
func To(name string, t any) *assocBuilder {
	return &assocBuilder{desc: &Descriptor{Name: name, Type: typ(t)}}
}

// From represents a reversed-edge between two vertices that has a back-reference to its source edge.
func From(name string, t any) *inverseBuilder {
	return &inverseBuilder{desc: &Descriptor{Name: name, Type: typ(t), Inverse: true}}
}

func typ(t any) string {
	if rt := reflect.TypeOf(t); rt.NumIn() > 0 {
		return rt.In(0).Name()
	}
	return ""
}

// assocBuilder is the builder for assoc edges.
type assocBuilder struct {
	desc *Descriptor
}

// Unique sets the edge type to be unique. Basically, it limits the edge to be one of the two:
// one2one or one2many. one2one applied if the inverse-edge is also unique.
func (b *assocBuilder) Unique() *assocBuilder {
	b.desc.Unique = true
	return b
}

// Required indicates that this edge is a required field on creation.
// Unlike fields, edges are optional by default.
func (b *assocBuilder) Required() *assocBuilder {
	b.desc.Required = true
	return b
}

// Immutable indicates that this edge cannot be updated.
func (b *assocBuilder) Immutable() *assocBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the assoc edge.
func (b *assocBuilder) StructTag(s string) *assocBuilder {
	b.desc.Tag = s
	return b
}

// From creates an inverse-edge with the same type.
func (b *assocBuilder) From(name string) *inverseBuilder {
	return &inverseBuilder{desc: &Descriptor{Name: name, Type: b.desc.Type, Inverse: true, Ref: b.desc}}
}

// Field is used to bind an edge (with a foreign-key) to a field in the schema.
//
//	field.Int("owner_id").
//		Optional()
//
//	edge.To("owner", User.Type).
//		Field("owner_id").
//		Unique(),
func (b *assocBuilder) Field(f string) *assocBuilder {
	b.desc.Field = f
	return b
}

// Through allows setting an "edge schema" to interact explicitly with M2M edges.
//
//	edge.To("friends", User.Type).
//		Through("friendships", Friendship.Type)
func (b *assocBuilder) Through(name string, t any) *assocBuilder {
	b.desc.Through = &struct{ N, T string }{N: name, T: typ(t)}
	return b
}

// Comment used to put annotations on the schema.
func (b *assocBuilder) Comment(c string) *assocBuilder {
	b.desc.Comment = c
	return b
}

// StorageKey sets the storage key of the edge.
//
//	edge.To("groups", Group.Type).
//		StorageKey(edge.Table("user_groups"), edge.Columns("user_id", "group_id"))
func (b *assocBuilder) StorageKey(opts ...StorageOption) *assocBuilder {
	if b.desc.StorageKey == nil {
		b.desc.StorageKey = &StorageKey{}
	}
	for i := range opts {
		opts[i](b.desc.StorageKey)
	}
	return b
}

// Annotations adds a list of annotations to the edge object to be used by
// codegen extensions.
//
//	edge.To("pets", Pet.Type).
//		Annotations(entgql.Bind())
func (b *assocBuilder) Annotations(annotations ...schema.Annotation) *assocBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Descriptor interface.
func (b *assocBuilder) Descriptor() *Descriptor {
	return b.desc
}

// inverseBuilder is the builder for inverse edges.
type inverseBuilder struct {
	desc *Descriptor
}

// Ref sets the referenced-edge of this inverse edge.
func (b *inverseBuilder) Ref(ref string) *inverseBuilder {
	b.desc.RefName = ref
	return b
}

// Unique sets the edge type to be unique. Basically, it limits the edge to be one of the two:
// one-2-one or one-2-many. one-2-one applied if the inverse-edge is also unique.
func (b *inverseBuilder) Unique() *inverseBuilder {
	b.desc.Unique = true
	return b
}

// Required indicates that this edge is a required field on creation.
// Unlike fields, edges are optional by default.
func (b *inverseBuilder) Required() *inverseBuilder {
	b.desc.Required = true
	return b
}

// Immutable indicates that this edge cannot be updated.
func (b *inverseBuilder) Immutable() *inverseBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the inverse edge.
func (b *inverseBuilder) StructTag(s string) *inverseBuilder {
	b.desc.Tag = s
	return b
}

// Comment used to put annotations on the schema.
func (b *inverseBuilder) Comment(c string) *inverseBuilder {
	b.desc.Comment = c
	return b
}

// Field is used to bind an edge (with a foreign-key) to a field in the schema.
//
//	field.Int("owner_id").
//		Optional()
//
//	edge.From("owner", User.Type).
//		Ref("pets").
//		Field("owner_id").
//		Unique(),
func (b *inverseBuilder) Field(f string) *inverseBuilder {
	b.desc.Field = f
	return b
}

// Through allows setting an "edge schema" to interact explicitly with M2M edges.
//
//	edge.From("liked_users", User.Type).
//		Ref("liked_tweets").
//		Through("likes", TweetLike.Type)
func (b *inverseBuilder) Through(name string, t any) *inverseBuilder {
	b.desc.Through = &struct{ N, T string }{N: name, T: typ(t)}
	return b
}

// Annotations adds a list of annotations to the edge object to be used by
// codegen extensions.
//
//	edge.From("owner", User.Type).
//		Ref("pets").
//		Unique().
//		Annotations(entgql.Bind())
func (b *inverseBuilder) Annotations(annotations ...schema.Annotation) *inverseBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Descriptor interface.
func (b *inverseBuilder) Descriptor() *Descriptor {
	return b.desc
}

// polymorphicTypeEntry represents a single type mapping in a polymorphic edge.
type polymorphicTypeEntry struct {
	TypeName  string
	TypeValue any
	FieldName string
}

// polymorphicAssocBuilder is the builder for polymorphic association edges using ToOneOf/FromOneOf pattern.
type polymorphicAssocBuilder struct {
	desc    *Descriptor
	entries []polymorphicTypeEntry
}

// ToOneOf creates a polymorphic association edge using type-value pairs.
//
//	edge.ToOneOf(
//		"message_type",
//		"text_message", TextMessage.Type,
//		"photo_message", PhotoMessage.Type,
//	)
func ToOneOf(discriminatorField string, typeValuePairs ...any) *polymorphicAssocBuilder {
	if len(typeValuePairs)%2 != 0 {
		panic("ToOneOf requires an even number of arguments (type-value pairs)")
	}

	entries := make([]polymorphicTypeEntry, 0, len(typeValuePairs)/2)
	typeNames := make([]string, 0, len(typeValuePairs)/2)

	for i := 0; i < len(typeValuePairs); i += 2 {
		typeName, ok := typeValuePairs[i].(string)
		if !ok {
			panic("ToOneOf: type name must be a string")
		}
		typeValue := typeValuePairs[i+1]

		entries = append(entries, polymorphicTypeEntry{
			TypeName:  typeName,
			TypeValue: typeValue,
			FieldName: typeName + "_id", // Default field name
		})
		typeNames = append(typeNames, typ(typeValue))
	}

	// Infer field name from discriminator field (remove _type suffix and add _id)
	fieldName := discriminatorField
	if strings.HasSuffix(fieldName, "_type") {
		fieldName = strings.TrimSuffix(fieldName, "_type") + "_id"
	} else {
		fieldName = fieldName + "_id"
	}

	return &polymorphicAssocBuilder{
		desc: &Descriptor{
			Name:                   discriminatorField,
			Type:                   "polymorphic",
			Field:                  fieldName,
			AllowedTypes:           typeNames,
			TypeDiscriminatorField: discriminatorField,
			IsPolymorphic:          true,
		},
		entries: entries,
	}
}

// FromOneOf creates a polymorphic inverse edge using type-value pairs.
//
//	edge.FromOneOf(
//		"owner_user_type",
//		"local_user", LocalUser.Type,
//		"foreign_user", ForeignUser.Type,
//	).Unique()
func FromOneOf(discriminatorField string, typeValuePairs ...any) *polymorphicInverseBuilder {
	if len(typeValuePairs)%2 != 0 {
		panic("FromOneOf requires an even number of arguments (type-value pairs)")
	}

	entries := make([]polymorphicTypeEntry, 0, len(typeValuePairs)/2)
	typeNames := make([]string, 0, len(typeValuePairs)/2)

	for i := 0; i < len(typeValuePairs); i += 2 {
		typeName, ok := typeValuePairs[i].(string)
		if !ok {
			panic("FromOneOf: type name must be a string")
		}
		typeValue := typeValuePairs[i+1]

		entries = append(entries, polymorphicTypeEntry{
			TypeName:  typeName,
			TypeValue: typeValue,
			FieldName: typeName + "_id", // Default field name
		})
		typeNames = append(typeNames, typ(typeValue))
	}

	// Infer field name from discriminator field (remove _type suffix and add _id)
	fieldName := discriminatorField
	if strings.HasSuffix(fieldName, "_type") {
		fieldName = strings.TrimSuffix(fieldName, "_type") + "_id"
	} else {
		fieldName = fieldName + "_id"
	}

	return &polymorphicInverseBuilder{
		desc: &Descriptor{
			Name:                   discriminatorField,
			Type:                   "polymorphic",
			Field:                  fieldName,
			AllowedTypes:           typeNames,
			TypeDiscriminatorField: discriminatorField,
			IsPolymorphic:          true,
			Inverse:                true,
		},
		entries: entries,
	}
}

// Required indicates that this edge is a required field on creation.
func (b *polymorphicAssocBuilder) Required() *polymorphicAssocBuilder {
	b.desc.Required = true
	return b
}

// Unique indicates that this edge is unique (creates a unique constraint).
func (b *polymorphicAssocBuilder) Unique() *polymorphicAssocBuilder {
	b.desc.Unique = true
	return b
}

// Immutable indicates that this edge cannot be updated.
func (b *polymorphicAssocBuilder) Immutable() *polymorphicAssocBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the polymorphic edge.
func (b *polymorphicAssocBuilder) StructTag(s string) *polymorphicAssocBuilder {
	b.desc.Tag = s
	return b
}

// Field sets the field that stores the polymorphic foreign key.
func (b *polymorphicAssocBuilder) Field(field string) *polymorphicAssocBuilder {
	b.desc.Field = field
	return b
}

// Comment sets the comment of the polymorphic edge.
func (b *polymorphicAssocBuilder) Comment(comment string) *polymorphicAssocBuilder {
	b.desc.Comment = comment
	return b
}

// Annotations adds annotations to the polymorphic edge.
func (b *polymorphicAssocBuilder) Annotations(annotations ...schema.Annotation) *polymorphicAssocBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor returns the edge descriptor.
func (b *polymorphicAssocBuilder) Descriptor() *Descriptor {
	return b.desc
}

// polymorphicInverseBuilder is the builder for polymorphic inverse edges using FromOneOf pattern.
type polymorphicInverseBuilder struct {
	desc    *Descriptor
	entries []polymorphicTypeEntry
}

// Ref sets the referenced-edge of this polymorphic inverse edge.
func (b *polymorphicInverseBuilder) Ref(ref string) *polymorphicInverseBuilder {
	b.desc.RefName = ref
	return b
}

// Required indicates that this edge is a required field on creation.
func (b *polymorphicInverseBuilder) Required() *polymorphicInverseBuilder {
	b.desc.Required = true
	return b
}

// Unique indicates that this edge is unique (creates a unique constraint).
func (b *polymorphicInverseBuilder) Unique() *polymorphicInverseBuilder {
	b.desc.Unique = true
	return b
}

// Immutable indicates that this edge cannot be updated.
func (b *polymorphicInverseBuilder) Immutable() *polymorphicInverseBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the polymorphic inverse edge.
func (b *polymorphicInverseBuilder) StructTag(s string) *polymorphicInverseBuilder {
	b.desc.Tag = s
	return b
}

// Field sets the field that stores the polymorphic foreign key.
func (b *polymorphicInverseBuilder) Field(field string) *polymorphicInverseBuilder {
	b.desc.Field = field
	return b
}

// Comment sets the comment of the polymorphic inverse edge.
func (b *polymorphicInverseBuilder) Comment(comment string) *polymorphicInverseBuilder {
	b.desc.Comment = comment
	return b
}

// Annotations adds annotations to the polymorphic inverse edge.
func (b *polymorphicInverseBuilder) Annotations(annotations ...schema.Annotation) *polymorphicInverseBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor returns the edge descriptor.
func (b *polymorphicInverseBuilder) Descriptor() *Descriptor {
	return b.desc
}

// StorageKey holds the configuration for edge storage-key.
type StorageKey struct {
	Table   string   // Table or label.
	Symbols []string // Symbols/names of the foreign-key constraints.
	Columns []string // Foreign-key columns.
}

// StorageOption allows for setting the storage configuration using functional options.
type StorageOption func(*StorageKey)

// Table sets the table name option for M2M edges.
func Table(name string) StorageOption {
	return func(key *StorageKey) {
		key.Table = name
	}
}

// Symbol sets the symbol/name of the foreign-key constraint for O2O, O2M and M2O edges.
// Note that, for M2M edges (2 columns and 2 constraints), use the edge.Symbols option.
func Symbol(symbol string) StorageOption {
	return func(key *StorageKey) {
		key.Symbols = []string{symbol}
	}
}

// Symbols sets the symbol/name of the foreign-key constraints for M2M edges.
// The 1st column defines the name of the "To" edge, and the 2nd defines
// the name of the "From" edge (inverse edge).
// Note that, for O2O, O2M and M2O edges, use the edge.Symbol option.
func Symbols(to, from string) StorageOption {
	return func(key *StorageKey) {
		key.Symbols = []string{to, from}
	}
}

// Column sets the foreign-key column name option for O2O, O2M and M2O edges.
// Note that, for M2M edges (2 columns), use the edge.Columns option.
func Column(name string) StorageOption {
	return func(key *StorageKey) {
		key.Columns = []string{name}
	}
}

// Columns sets the foreign-key column names option for M2M edges.
// The 1st column defines the name of the "To" edge, and the 2nd defines
// the name of the "From" edge (inverse edge).
// Note that, for O2O, O2M and M2O edges, use the edge.Column option.
func Columns(to, from string) StorageOption {
	return func(key *StorageKey) {
		key.Columns = []string{to, from}
	}
}
