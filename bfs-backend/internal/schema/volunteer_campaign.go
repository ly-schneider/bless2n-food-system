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

type VolunteerCampaign struct {
	ent.Schema
}

func (VolunteerCampaign) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "volunteer_campaign"},
	}
}

func (VolunteerCampaign) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.UUID("claim_token", uuid.UUID{}).
			Default(uuidV7).
			Unique(),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("access_code").
			MaxLen(4),
		field.Time("valid_from").
			Optional().
			Nillable(),
		field.Time("valid_until").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("draft", "active", "ended").
			Default("active").
			StorageKey("status"),
		field.Int("max_redemptions").
			Default(1).
			Positive(),
		field.Int("redemption_count").
			Default(0).
			NonNegative(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (VolunteerCampaign) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products", Product.Type).
			Through("campaign_products", VolunteerCampaignProduct.Type),
		edge.To("redemptions", VolunteerRedemption.Type),
	}
}

func (VolunteerCampaign) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("claim_token").Unique(),
		index.Fields("status"),
	}
}
