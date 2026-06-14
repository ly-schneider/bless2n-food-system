package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OrderLineRedemption struct {
	ent.Schema
}

func (OrderLineRedemption) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "order_line_redemption"},
	}
}

func (OrderLineRedemption) Fields() []ent.Field {
	return []ent.Field{
		nanoidPK(),
		field.String("order_line_id").
			MaxLen(36).
			NotEmpty().
			Unique(),
		field.Time("redeemed_at").
			Default(time.Now),
	}
}

func (OrderLineRedemption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order_line", OrderLine.Type).
			Ref("redemption").
			Field("order_line_id").
			Unique().
			Required(),
	}
}
