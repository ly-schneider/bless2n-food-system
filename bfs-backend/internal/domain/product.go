package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductType string

const (
	ProductTypeSimple ProductType = "simple"
	ProductTypeMenu   ProductType = "menu"
)

type Product struct {
	ID         bson.ObjectID `bson:"_id"`
	CategoryID bson.ObjectID `bson:"category_id" validate:"required"`
	Type       ProductType   `bson:"type" validate:"required,oneof=simple menu"`
	Name       string        `bson:"name" validate:"required"`
	Image      *string       `bson:"image,omitempty"`
	PriceCents Cents         `bson:"price_cents" validate:"required,gte=0"`
	IsActive   bool          `bson:"is_active"`
	CreatedAt  time.Time     `bson:"created_at"`
	UpdatedAt  time.Time     `bson:"updated_at"`
}

type ProductSummaryDTO struct {
	ID         string      `json:"id"`
	Category   CategoryDTO `json:"category"`
	Type       ProductType `json:"type"`
	Name       string      `json:"name"`
	Image      *string     `json:"image,omitempty"`
	PriceCents Cents       `json:"priceCents"`
	IsActive   bool        `json:"isActive"`
	// Inventory snapshot for availability on menus
	AvailableQuantity *int  `json:"availableQuantity,omitempty"`
	IsAvailable       *bool `json:"isAvailable,omitempty"`
	IsLowStock        *bool `json:"isLowStock,omitempty"`
}

type ProductDTO struct {
	ProductSummaryDTO
	Menu *MenuDTO `json:"menu,omitempty"`
}
