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

type Device struct {
	ent.Schema
}

func (Device) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device"},
	}
}

func (Device) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuidV7).
			Immutable(),
		field.String("name").
			MaxLen(20).
			NotEmpty(),
		field.String("model").
			MaxLen(50).
			Optional().
			Nillable(),
		field.String("os").
			MaxLen(20).
			Optional().
			Nillable(),
		field.String("device_key").
			MaxLen(100).
			NotEmpty().
			Unique(),
		field.Enum("type").
			Values("POS", "STATION").
			StorageKey("type"),
		field.Enum("status").
			Values("pending", "approved", "rejected", "revoked").
			Default("pending").
			StorageKey("status"),
		field.String("decided_by").
			Optional().
			Nillable(),
		field.Time("decided_at").
			Optional().
			Nillable(),
		field.Time("expires_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("pending_session_token").
			Optional().
			Nillable(),
		field.String("pairing_code").
			MaxLen(6).
			Optional().
			Nillable(),
		field.Time("pairing_code_expires_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Device.
func (Device) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products", Product.Type).
			Through("device_products", DeviceProduct.Type),
		edge.To("order_payments", OrderPayment.Type),
		edge.To("inventory_ledger_entries", InventoryLedger.Type),
	}
}
