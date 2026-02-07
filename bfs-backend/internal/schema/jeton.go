package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Jeton struct {
	ent.Schema
}

func (Jeton) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "jeton"},
	}
}

func (Jeton) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.String("name").
			MaxLen(20).
			NotEmpty(),
		field.String("color").
			MaxLen(7).
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Jeton) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products", Product.Type),
	}
}
