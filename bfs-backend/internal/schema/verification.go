package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Verification struct {
	ent.Schema
}

func (Verification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "verification"},
	}
}

func (Verification) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable(),
		field.String("identifier").
			NotEmpty(),
		field.String("value").
			NotEmpty(),
		field.Time("expires_at"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Verification) Edges() []ent.Edge {
	return nil
}
