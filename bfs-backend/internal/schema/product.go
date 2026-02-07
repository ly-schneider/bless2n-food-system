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
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.UUID("category_id", uuid.UUID{}),
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
		field.Int64("price_cents").
			Default(0),
		field.UUID("jeton_id", uuid.UUID{}).
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
	}
}
