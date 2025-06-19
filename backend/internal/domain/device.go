package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type Device struct {
	ID           model.NanoID14 `gorm:"type:nano_id;primaryKey" json:"id"`
	SerialNumber string         `gorm:"unique;not null"         json:"serial_number" validate:"required"`
	Model        string         `gorm:"not null"                json:"model"          validate:"required"`
	IsActive     bool           `gorm:"not null;default:true"   json:"is_active"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"          json:"created_at"`
	UpdatedAt    *time.Time     `gorm:"autoUpdateTime"          json:"updated_at,omitempty"`
}

func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = model.NanoID14(utils.Must())
	}
	return nil
}
