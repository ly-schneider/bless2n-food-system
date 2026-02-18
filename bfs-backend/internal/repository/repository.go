package repository

import (
	"backend/internal/generated/ent"
)

type Repositories struct {
	User    UserRepository
	Order   OrderRepository
	Product *ProductRepository
}

func New(client *ent.Client) *Repositories {
	return &Repositories{
		User:    NewUserRepository(client),
		Order:   NewOrderRepository(client),
		Product: NewProductRepository(client),
	}
}
