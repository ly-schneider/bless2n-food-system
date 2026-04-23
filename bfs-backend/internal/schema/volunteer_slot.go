package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type VolunteerSlot struct {
	ent.Schema
}

func (VolunteerSlot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "volunteer_slot"},
	}
}

func (VolunteerSlot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.UUID("campaign_id", uuid.UUID{}),
		field.UUID("order_id", uuid.UUID{}).
			Unique(),
		field.String("reserved_by_session").
			MaxLen(64).
			Optional().
			Nillable(),
		field.Time("reserved_at").
			Optional().
			Nillable(),
		field.Time("reserved_until").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (VolunteerSlot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", VolunteerCampaign.Type).
			Ref("slots").
			Field("campaign_id").
			Unique().
			Required(),
		edge.To("order", Order.Type).
			Field("order_id").
			Unique().
			Required(),
	}
}

func (VolunteerSlot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id", "reserved_until"),
		index.Fields("reserved_by_session"),
	}
}
