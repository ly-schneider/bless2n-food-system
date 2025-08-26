package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID             model.NanoID14 `gorm:"type:nano_id;primaryKey"               json:"id"`
	OwnerID        model.NanoID14 `gorm:"type:nano_id;not null"                 json:"owner_id"`
	Owner          *User          `gorm:"constraint:OnDelete:RESTRICT"          json:"owner,omitempty"`
	Name           string         `gorm:"not null"                              json:"name"           validate:"required"`
	Location       string         `gorm:"not null"                              json:"location"       validate:"required"`
	CheckoutSpots  int            `gorm:"not null;check:checkout_spots >= 0"    json:"checkout_spots"`
	IsSelfCheckout bool           `gorm:"not null;default:false"                json:"is_self_checkout"`
	StartDate      *time.Time     `gorm:"type:date"                             json:"start_date,omitempty"`
	EndDate        *time.Time     `gorm:"type:date"                             json:"end_date,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                        json:"created_at"`
	UpdatedAt      *time.Time     `gorm:"autoUpdateTime"                        json:"updated_at,omitempty"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = model.NanoID14(utils.Must())
	}
	return nil
}
