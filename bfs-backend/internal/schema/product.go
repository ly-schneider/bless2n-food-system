package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Product struct {
	ent.Schema
}

func (Product) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "product"},
	}
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		nanoidPK(),
		field.String("category_id").
			MaxLen(36),
		field.Enum("type").
			Values("simple", "menu").
			Default("simple").
			StorageKey("type"),
		field.String("name").
			MaxLen(20).
			NotEmpty(),
		field.String("image").
			Optional().
			Nillable(),
		field.String("description").
			MaxLen(500).
			Optional().
			Nillable(),
		field.Int64("price_cents").
			Default(0),
		field.String("jeton_id").
			MaxLen(36).
			Optional().
			Nillable(),
		field.Bool("is_active").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", Category.Type).
			Ref("products").
			Field("category_id").
			Unique().
			Required(),
		edge.From("jeton", Jeton.Type).
			Ref("products").
			Field("jeton_id").
			Unique(),
		edge.To("menu_slots", MenuSlot.Type),
		edge.From("menu_slot_options", MenuSlot.Type).
			Ref("option_products").
			Through("menu_slot_option_links", MenuSlotOption.Type),
		edge.From("devices", Device.Type).
			Ref("products").
			Through("device_products", DeviceProduct.Type),
		edge.To("order_lines", OrderLine.Type),
		edge.To("inventory_ledger_entries", InventoryLedger.Type),
		edge.From("club100_settings", Settings.Type).
			Ref("club100_free_products").
			Through("club100_free_product_links", Club100FreeProduct.Type),
		edge.From("volunteer_campaigns", VolunteerCampaign.Type).
			Ref("products").
			Through("campaign_products", VolunteerCampaignProduct.Type),
	}
}
