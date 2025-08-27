package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProductBundleComponent struct {
	BundleID           primitive.ObjectID `bson:"bundle_id" json:"bundle_id" validate:"required"`
	ComponentProductID primitive.ObjectID `bson:"component_product_id" json:"component_product_id" validate:"required"`
	Quantity           int                `bson:"qty" json:"qty" validate:"required,gt=0"`
}