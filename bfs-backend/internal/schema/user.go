package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user"},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable(),
		field.String("name").
			Optional().
			Nillable(),
		field.String("email").
			Optional().
			Nillable().
			Unique(),
		field.Bool("email_verified").
			Default(false),
		field.String("image").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Bool("is_anonymous").
			Default(false),
		field.Enum("role").
			Values("admin", "customer").
			Default("customer").
			StorageKey("role"),
		field.Bool("is_club_100").
			Default(false),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", Session.Type),
		edge.To("invites", AdminInvite.Type),
	}
}
