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
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.UUID("product_id", uuid.UUID{}),
		field.Int("delta"),
		field.Enum("reason").
			Values("opening_balance", "sale", "refund", "cancellation", "manual_adjust", "correction").
			StorageKey("reason"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.UUID("order_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("order_line_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("device_id", uuid.UUID{}).
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
