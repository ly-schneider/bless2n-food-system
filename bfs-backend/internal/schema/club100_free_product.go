package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Club100FreeProduct struct {
	ent.Schema
}

func (Club100FreeProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "club100_free_product"},
		field.ID("settings_id", "product_id"),
	}
}

func (Club100FreeProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("settings_id").MaxLen(50),
		field.UUID("product_id", uuid.UUID{}),
	}
}

func (Club100FreeProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("settings", Settings.Type).
			Field("settings_id").
			Unique().
			Required(),
		edge.To("product", Product.Type).
			Field("product_id").
			Unique().
			Required(),
	}
}
