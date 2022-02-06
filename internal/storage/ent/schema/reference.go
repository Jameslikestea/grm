package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Reference holds the schema definition for the Reference entity.
type Reference struct {
	ent.Schema
}

// Fields of the Reference.
func (Reference) Fields() []ent.Field {
	return []ent.Field{
		field.String("package"),
		field.String("ref"),
		field.String("hash"),
	}
}

// Edges of the Reference.
func (Reference) Edges() []ent.Edge {
	return nil
}

func (Reference) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("package", "ref").Unique(),
	}
}
