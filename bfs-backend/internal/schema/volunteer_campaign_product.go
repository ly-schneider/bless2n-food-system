package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type VolunteerCampaignProduct struct {
	ent.Schema
}

func (VolunteerCampaignProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "volunteer_campaign_product"},
		field.ID("campaign_id", "product_id"),
	}
}

func (VolunteerCampaignProduct) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("campaign_id", uuid.UUID{}),
		field.UUID("product_id", uuid.UUID{}),
		field.Int("quantity").
			Default(1),
	}
}

func (VolunteerCampaignProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("campaign", VolunteerCampaign.Type).
			Field("campaign_id").
			Unique().
			Required(),
		edge.To("product", Product.Type).
			Field("product_id").
			Unique().
			Required(),
	}
}
