package model

import (
	"github.com/google/uuid"
)

// MenuSlotOption represents a product option available in a menu slot.
type MenuSlotOption struct {
	MenuSlotID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	OptionProductID uuid.UUID `gorm:"type:uuid;primaryKey"`

	// Relations
	MenuSlot      MenuSlot `gorm:"foreignKey:MenuSlotID"`
	OptionProduct Product  `gorm:"foreignKey:OptionProductID"`
}

func (MenuSlotOption) TableName() string {
	return "app.menu_slot_option"
}
