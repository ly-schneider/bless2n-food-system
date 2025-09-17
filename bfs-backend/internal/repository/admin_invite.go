package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type AdminInviteRepository interface {
}

type adminInviteRepository struct {
	collection *mongo.Collection
}

func NewAdminInviteRepository(db *database.MongoDB) AdminInviteRepository {
	return &adminInviteRepository{
		collection: db.Database.Collection(database.AdminInvitesCollection),
	}
}
