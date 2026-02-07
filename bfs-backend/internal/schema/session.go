package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Session struct {
	ent.Schema
}

func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "session"},
	}
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable(),
		field.String("token").
			NotEmpty().
			Unique(),
		field.String("user_id").
			NotEmpty(),
		field.Time("expires_at"),
		field.String("ip_address").
			Optional().
			Nillable(),
		field.String("user_agent").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sessions").
			Field("user_id").
			Unique().
			Required(),
	}
}
