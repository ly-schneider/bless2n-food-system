package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Idempotency struct {
	ent.Schema
}

func (Idempotency) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "idempotency"},
	}
}

func (Idempotency) Fields() []ent.Field {
	return []ent.Field{
		nanoidPK(),
		field.String("scope").
			MaxLen(100).
			NotEmpty(),
		field.String("key").
			MaxLen(100).
			NotEmpty(),
		field.JSON("response", []byte{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("expires_at"),
	}
}

func (Idempotency) Edges() []ent.Edge {
	return nil
}

func (Idempotency) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("scope", "key").Unique(),
		index.Fields("expires_at"),
	}
}
