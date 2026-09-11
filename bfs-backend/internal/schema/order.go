package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Order struct {
	ent.Schema
}

func (Order) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "order"},
	}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		nanoidPK(),
		field.String("customer_id").
			Optional().
			Nillable(),
		field.String("contact_email").
			MaxLen(50).
			Optional().
			Nillable(),
		field.Int64("total_cents"),
		field.Enum("status").
			Values("pending", "paid", "cancelled", "refunded").
			StorageKey("status"),
		field.Enum("origin").
			Values("shop", "pos").
			StorageKey("origin"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("payment_attempt_id").
			MaxLen(100).
			Optional().
			Nillable(),
		field.Int("payrexx_gateway_id").
			Optional().
			Nillable(),
		field.Int("payrexx_transaction_id").
			Optional().
			Nillable(),
		// Signed pickup token (see internal/qrsign), persisted so reprints don't re-sign.
		field.Text("qr_payload").
			Optional().
			Nillable(),
	}
}

func (Order) Edges() []ent.Edge {
	// ON DELETE mirrors db/migrations; Ent reads it from this assoc (To) side,
	// not the child's inverse edge.
	return []ent.Edge{
		edge.To("payments", OrderPayment.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("lines", OrderLine.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("inventory_ledger_entries", InventoryLedger.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("club100_redemptions", Club100Redemption.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
