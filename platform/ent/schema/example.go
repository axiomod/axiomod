// Package schema holds the Ent schema definitions for the framework's
// reference domain. Run `go generate ./platform/ent/...` after changing it.
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Example holds the schema definition for the Example entity. It mirrors
// examples/example/entity.Example: the domain entity stays the source of
// truth and the Ent model is a persistence detail behind the repository.
type Example struct {
	ent.Schema
}

// Fields of the Example.
func (Example) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable().
			Comment("UUID assigned by the domain entity constructor"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			NotEmpty(),
		field.String("value_type").
			Default(""),
		field.Int("value_count").
			Default(0),
		field.JSON("value_tags", []string{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Example.
func (Example) Edges() []ent.Edge {
	return nil
}
