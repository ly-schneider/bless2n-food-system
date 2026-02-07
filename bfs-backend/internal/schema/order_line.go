package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type OrderLine struct {
	ent.Schema
}

func (OrderLine) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "order_line"},
	}
}

func (OrderLine) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.UUID("order_id", uuid.UUID{}),
		field.Enum("line_type").
			Values("simple", "bundle", "component").
			StorageKey("line_type"),
		field.UUID("product_id", uuid.UUID{}),
		field.String("title").
			MaxLen(20).
			NotEmpty(),
		field.Int("quantity").
			Default(1),
		field.Int64("unit_price_cents").
			Default(0),
		field.UUID("parent_line_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("menu_slot_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("menu_slot_name").
			MaxLen(20).
			Optional().
			Nillable(),
	}
}

func (OrderLine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("lines").
			Field("order_id").
			Unique().
			Required(),
		edge.From("product", Product.Type).
			Ref("order_lines").
			Field("product_id").
			Unique().
			Required(),
		// Self-referencing: parent/child lines
		edge.To("child_lines", OrderLine.Type),
		edge.From("parent_line", OrderLine.Type).
			Ref("child_lines").
			Field("parent_line_id").
			Unique(),
		edge.From("menu_slot", MenuSlot.Type).
			Ref("order_lines").
			Field("menu_slot_id").
			Unique(),
		edge.To("redemption", OrderLineRedemption.Type).
			Unique(),
		edge.To("inventory_ledger_entries", InventoryLedger.Type),
	}
}
