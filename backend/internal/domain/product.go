package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductType string

const (
	ProductTypeSimple ProductType = "simple"
	ProductTypeBundle ProductType = "bundle"
)

type Product struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CategoryID primitive.ObjectID `bson:"category_id" json:"category_id" validate:"required"`
	Type       ProductType        `bson:"type" json:"type" validate:"required,oneof=simple bundle"`
	Name       string             `bson:"name" json:"name" validate:"required"`
	Image      *string            `bson:"image,omitempty" json:"image,omitempty"`
	Price      float64            `bson:"price" json:"price" validate:"required,gte=0"`
	IsActive   bool               `bson:"is_active" json:"is_active"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}