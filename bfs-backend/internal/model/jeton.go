package model

import (
	"time"

	"github.com/google/uuid"
)

// Jeton represents a jeton (token) used for order fulfillment.
type Jeton struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string    `gorm:"type:varchar(20);not null"`
	PaletteColor string    `gorm:"type:varchar(20);not null"`
	HexColor     *string   `gorm:"type:varchar(7)"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"not null;autoUpdateTime"`

	// Relations
	Products []Product `gorm:"foreignKey:JetonID"`
}

func (Jeton) TableName() string {
	return "app.jeton"
}
