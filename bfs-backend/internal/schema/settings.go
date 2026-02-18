package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Settings struct {
	ent.Schema
}

func (Settings) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "settings"},
	}
}

func (Settings) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(50).
			Default("default").
			Immutable(),
		field.Enum("pos_mode").
			Values("QR_CODE", "JETON").
			Default("QR_CODE").
			StorageKey("pos_mode"),
		field.Int("club100_max_redemptions").
			Default(2),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Settings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("club100_free_products", Product.Type).
			Through("club100_free_product_links", Club100FreeProduct.Type),
	}
}
