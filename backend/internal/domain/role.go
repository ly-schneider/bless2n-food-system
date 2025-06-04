package domain

type Role struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name" validate:"required"`
}

// Create a map of roles for easy reference
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
