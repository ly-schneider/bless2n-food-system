package model

import (
	"github.com/google/uuid"
)

// MenuSlot represents a slot in a menu product that can contain options.
type MenuSlot struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MenuProductID uuid.UUID `gorm:"type:uuid;not null"`
	Name          string    `gorm:"type:varchar(20);not null"`
	Sequence      int       `gorm:"not null;default:0"`

	// Relations
	MenuProduct Product          `gorm:"foreignKey:MenuProductID"`
	Options     []MenuSlotOption `gorm:"foreignKey:MenuSlotID"`
}

func (MenuSlot) TableName() string {
	return "app.menu_slot"
}
