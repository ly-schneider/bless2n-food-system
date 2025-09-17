package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type InventoryLedgerRepository interface {
}

type inventoryLedgerRepository struct {
	collection *mongo.Collection
}

func NewInventoryLedgerRepository(db *database.MongoDB) InventoryLedgerRepository {
	return &inventoryLedgerRepository{
		collection: db.Database.Collection(database.InventoryLedgerCollection),
	}
}
