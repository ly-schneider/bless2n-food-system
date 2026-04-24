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

type VolunteerRedemption struct {
	ent.Schema
}

func (VolunteerRedemption) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "volunteer_redemption"},
	}
}

func (VolunteerRedemption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.UUID("campaign_id", uuid.UUID{}),
		field.UUID("order_id", uuid.UUID{}).
			Unique(),
		field.UUID("station_device_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("idempotency_key").
			MaxLen(64).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (VolunteerRedemption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", VolunteerCampaign.Type).
			Ref("redemptions").
			Field("campaign_id").
			Unique().
			Required(),
		edge.To("order", Order.Type).
			Field("order_id").
			Unique().
			Required(),
	}
}

func (VolunteerRedemption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
