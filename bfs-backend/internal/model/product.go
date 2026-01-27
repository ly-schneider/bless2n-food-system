package model

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a purchasable product.
type Product struct {
	ID         uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CategoryID uuid.UUID   `gorm:"type:uuid;not null"`
	Type       ProductType `gorm:"type:product_type;not null;default:'simple'"`
	Name       string      `gorm:"type:varchar(20);not null"`
	Image      *string     `gorm:"type:text"`
	PriceCents int64       `gorm:"not null;default:0"`
	JetonID    *uuid.UUID  `gorm:"type:uuid"`
	IsActive   bool        `gorm:"not null;default:true"`
	CreatedAt  time.Time   `gorm:"not null;autoCreateTime"`
	UpdatedAt  time.Time   `gorm:"not null;autoUpdateTime"`

	// Relations
	Category  Category   `gorm:"foreignKey:CategoryID"`
	Jeton     *Jeton     `gorm:"foreignKey:JetonID"`
	MenuSlots []MenuSlot `gorm:"foreignKey:MenuProductID"`
}

func (Product) TableName() string {
	return "app.product"
}
