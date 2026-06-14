package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type DeviceProduct struct {
	ent.Schema
}

func (DeviceProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device_product"},
		field.ID("device_id", "product_id"),
	}
}

func (DeviceProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("device_id").
			MaxLen(36).
			NotEmpty(),
		field.String("product_id").
			MaxLen(36).
			NotEmpty(),
	}
}

func (DeviceProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("device", Device.Type).
			Field("device_id").
			Unique().
			Required(),
		edge.To("product", Product.Type).
			Field("product_id").
			Unique().
			Required(),
	}
}
