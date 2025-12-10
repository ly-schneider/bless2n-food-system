package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Jeton struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Name         string        `bson:"name" validate:"required"`
	PaletteColor string        `bson:"palette_color" validate:"required"`
	HexColor     *string       `bson:"hex_color,omitempty"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

type JetonDTO struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	PaletteColor string  `json:"paletteColor"`
	HexColor     *string `json:"hexColor,omitempty"`
	ColorHex     string  `json:"colorHex"`
	UsageCount   *int64  `json:"usageCount,omitempty"`
}
