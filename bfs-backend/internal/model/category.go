package model

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a product category.
type Category struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(20);not null"`
	IsActive  bool      `gorm:"not null;default:true"`
	Position  int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`

	// Relations
	Products []Product `gorm:"foreignKey:CategoryID"`
}

func (Category) TableName() string {
	return "app.category"
}
