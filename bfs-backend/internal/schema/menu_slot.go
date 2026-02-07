package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type MenuSlot struct {
	ent.Schema
}

func (MenuSlot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "menu_slot"},
	}
}

func (MenuSlot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.UUID("menu_product_id", uuid.UUID{}),
		field.String("name").
			MaxLen(20).
			NotEmpty(),
		field.Int("sequence").
			Default(0),
	}
}

func (MenuSlot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("menu_product", Product.Type).
			Ref("menu_slots").
			Field("menu_product_id").
			Unique().
			Required(),
		edge.To("option_products", Product.Type).
			Through("options", MenuSlotOption.Type),
		edge.To("order_lines", OrderLine.Type),
	}
}
