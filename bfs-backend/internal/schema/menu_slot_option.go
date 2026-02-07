package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type MenuSlotOption struct {
	ent.Schema
}

func (MenuSlotOption) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "menu_slot_option"},
		field.ID("menu_slot_id", "option_product_id"),
	}
}

func (MenuSlotOption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("menu_slot_id", uuid.UUID{}),
		field.UUID("option_product_id", uuid.UUID{}),
	}
}

func (MenuSlotOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("menu_slot", MenuSlot.Type).
			Field("menu_slot_id").
			Unique().
			Required(),
		edge.To("option_product", Product.Type).
			Field("option_product_id").
			Unique().
			Required(),
	}
}
