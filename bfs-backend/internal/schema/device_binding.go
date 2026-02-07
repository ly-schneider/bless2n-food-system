package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type DeviceBinding struct {
	ent.Schema
}

func (DeviceBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device_binding"},
	}
}

func (DeviceBinding) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.Enum("device_type").
			Values("POS", "STATION").
			StorageKey("device_type"),
		field.String("token_hash").
			NotEmpty().
			Unique(),
		field.String("name").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("last_seen_at").
			Default(time.Now),
		field.String("created_by_user_id").
			NotEmpty(),
		field.Time("revoked_at").
			Optional().
			Nillable(),
		field.UUID("station_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("device_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (DeviceBinding) Edges() []ent.Edge {
	return nil
}
