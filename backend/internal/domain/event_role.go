package domain

import (
	"time"
)

type EventRole struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"not null"                 json:"name"         validate:"required"`
	DisplayName string    `gorm:"not null"                 json:"display_name" validate:"required"`
	Description *string   `json:"description,omitempty"`
	IsDefault   bool      `gorm:"not null;default:false"   json:"is_default"`
	IsActive    bool      `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime"          json:"created_at"`
}

var EventRoles = map[string]Role{
	"admin": {
		ID:   1,
		Name: "admin",
	},
	"moderator": {
		ID:   2,
		Name: "moderator",
	},
	"member": {
		ID:   3,
		Name: "member",
	},
}
