package domain

import "time"

type Role struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"not null"                 json:"name"         validate:"required"`
	DisplayName string     `gorm:"not null"                 json:"display_name" validate:"required"`
	Description *string    `json:"description,omitempty"`
	IsDefault   bool       `gorm:"not null;default:false"   json:"is_default"`
	IsActive    bool       `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"          json:"created_at"`
	UpdatedAt   *time.Time `gorm:"autoUpdateTime"          json:"updated_at,omitempty"`
}

var Roles = map[string]Role{
	"admin": {
		ID:   1,
		Name: "admin",
	},
	"user": {
		ID:   2,
		Name: "user",
	},
}
