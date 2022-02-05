package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Object holds the schema definition for the Object entity.
type Object struct {
	ent.Schema
}

// CREATE TABLE IF NOT EXISTS objs (" +
//			"package TEXT," +
//			"type TINYINT," +
//			"hash TEXT," +
//			"content BLOB," +
//			"PRIMARY KEY (package, hash));",

// Fields of the Object.
func (Object) Fields() []ent.Field {
	return []ent.Field{
		field.String("package"),
		field.Int8("type"),
		field.String("hash"),
		field.Bytes("content"),
	}
}

// Edges of the Object.
func (Object) Edges() []ent.Edge {
	return nil
}

func (Object) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("package", "hash").Unique(),
	}
}
