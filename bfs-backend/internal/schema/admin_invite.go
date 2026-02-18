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

type AdminInvite struct {
	ent.Schema
}

func (AdminInvite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "admin_invite"},
	}
}

func (AdminInvite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.String("invited_by_user_id").
			NotEmpty().
			StorageKey("invited_by_user_id"),
		field.String("invitee_email").
			NotEmpty().
			StorageKey("invitee_email"),
		field.String("token_hash").
			NotEmpty().
			Unique().
			StorageKey("token_hash"),
		field.Enum("status").
			Values("pending", "accepted", "expired", "revoked").
			Default("pending").
			StorageKey("status"),
		field.Time("expires_at").
			StorageKey("expires_at"),
		field.Time("used_at").
			Optional().
			Nillable().
			StorageKey("used_at"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			StorageKey("created_at"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			StorageKey("updated_at"),
	}
}

func (AdminInvite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("inviter", User.Type).
			Ref("invites").
			Field("invited_by_user_id").
			Unique().
			Required(),
	}
}
