package model

import (
	"time"
)

// PosSettings represents global POS settings.
type PosSettings struct {
	ID        string             `gorm:"type:varchar(50);primaryKey;default:'default'"`
	Mode      PosFulfillmentMode `gorm:"type:pos_fulfillment_mode;not null;default:'JETON'"`
	UpdatedAt time.Time          `gorm:"not null;autoUpdateTime"`
}

func (PosSettings) TableName() string {
	return "app.pos_settings"
}
