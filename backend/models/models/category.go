package models

import "backend/models/abstract"

type Category struct {
	abstract.Base
	Name     string `gorm:"not null" json:"name" validate:"required"`
	Emoji    string `gorm:"not null" json:"emoji" validate:"required"`
	IsActive bool   `gorm:"default:true;not null" json:"is_active"`
}
