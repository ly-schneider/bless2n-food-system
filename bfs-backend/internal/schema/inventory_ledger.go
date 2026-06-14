package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type InventoryLedger struct {
	ent.Schema
}

func (InventoryLedger) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_ledger"},
	}
}

func (InventoryLedger) Fields() []ent.Field {
	return []ent.Field{
		nanoidPK(),
		field.String("product_id").
			MaxLen(36),
		field.Int("delta"),
		field.Enum("reason").
			Values("opening_balance", "sale", "refund", "cancellation", "manual_adjust", "correction").
			StorageKey("reason"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.String("order_id").
			MaxLen(36).
			Optional().
			Nillable(),
		field.String("order_line_id").
			MaxLen(36).
			Optional().
			Nillable(),
		field.String("device_id").
			MaxLen(36).
			Optional().
			Nillable(),
		field.String("created_by").
			Optional().
			Nillable(),
	}
}

func (InventoryLedger) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("inventory_ledger_entries").
			Field("product_id").
			Unique().
			Required(),
		edge.From("order", Order.Type).
			Ref("inventory_ledger_entries").
			Field("order_id").
			Unique(),
		edge.From("order_line", OrderLine.Type).
			Ref("inventory_ledger_entries").
			Field("order_line_id").
			Unique(),
		edge.From("device", Device.Type).
			Ref("inventory_ledger_entries").
			Field("device_id").
			Unique(),
	}
}
