package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OrderPayment struct {
	ent.Schema
}

func (OrderPayment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "order_payment"},
	}
}

func (OrderPayment) Fields() []ent.Field {
	return []ent.Field{
		nanoidPK(),
		field.String("order_id").
			MaxLen(36).
			NotEmpty(),
		field.Enum("method").
			Values("CASH", "CARD", "TWINT", "GRATIS_GUEST", "GRATIS_VIP", "GRATIS_STAFF", "GRATIS_100CLUB").
			StorageKey("method"),
		field.Int64("amount_cents"),
		field.String("device_id").
			MaxLen(36).
			Optional().
			Nillable(),
		field.Time("paid_at").
			Default(time.Now),
		field.String("card_brand").
			Optional().
			Nillable(),
		field.String("card_last4").
			Optional().
			Nillable(),
		field.String("entry_mode").
			Optional().
			Nillable(),
		field.String("card_transaction_id").
			Optional().
			Nillable(),
	}
}

func (OrderPayment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("payments").
			Field("order_id").
			Unique().
			Required(),
		edge.From("device", Device.Type).
			Ref("order_payments").
			Field("device_id").
			Unique(),
	}
}
