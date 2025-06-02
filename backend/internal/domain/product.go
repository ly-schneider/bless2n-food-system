package domain

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"size:255;not null"`
	Price     float64 `gorm:"type:numeric(10,2)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // enables soft-delete
}
