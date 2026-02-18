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

type Club100Redemption struct {
	ent.Schema
}

func (Club100Redemption) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "club100_redemption"},
	}
}

func (Club100Redemption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.String("elvanto_person_id").
			MaxLen(50).
			NotEmpty(),
		field.String("elvanto_person_name").
			MaxLen(100).
			NotEmpty(),
		field.UUID("order_id", uuid.UUID{}),
		field.Int("free_product_quantity").
			Default(1),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (Club100Redemption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("club100_redemptions").
			Field("order_id").
			Unique().
			Required(),
	}
}
