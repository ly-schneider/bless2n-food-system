package seed

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type UserSeeder struct {
	Logger *zap.Logger
}

type UserDocument struct {
	ID               bson.ObjectID   `bson:"_id"`
	Email            string          `bson:"email"`
	FirstName        string          `bson:"first_name,omitempty"`
	LastName         string          `bson:"last_name,omitempty"`
	Role             domain.UserRole `bson:"role"`
	IsVerified       bool            `bson:"is_verified"`
	StripeCustomerID *string         `bson:"stripe_customer_id,omitempty"`
	CreatedAt        time.Time       `bson:"created_at"`
	UpdatedAt        time.Time       `bson:"updated_at"`
}

type userSeed struct {
	Email      string
	FirstName  string
	LastName   string
	Role       domain.UserRole
	IsVerified bool
}

var userSeeds = []userSeed{
	{Email: "levyn.schneider@leys.ch", FirstName: "Levyn", LastName: "Schneider", Role: domain.UserRoleAdmin, IsVerified: true},
}

func NewUserSeeder(logger *zap.Logger) UserSeeder {
	return UserSeeder{Logger: logger}
}

func (s UserSeeder) Name() string {
	return "users"
}

func (s UserSeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	coll := db.Collection(database.UsersCollection)

	for _, seed := range userSeeds {
		now := time.Now().UTC()
		email := normalizeEmail(seed.Email)
		doc := UserDocument{
			ID:         bson.NewObjectID(),
			Email:      email,
			FirstName:  seed.FirstName,
			LastName:   seed.LastName,
			Role:       seed.Role,
			IsVerified: seed.IsVerified,
			CreatedAt:  seededAt,
			UpdatedAt:  now,
		}

		filter := bson.M{"email": doc.Email}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":        doc.ID,
				"created_at": doc.CreatedAt,
			},
			"$set": bson.M{
				"email":       doc.Email,
				"first_name":  doc.FirstName,
				"last_name":   doc.LastName,
				"role":        doc.Role,
				"is_verified": doc.IsVerified,
				"updated_at":  doc.UpdatedAt,
			},
		}

		opts := options.UpdateOne().SetUpsert(true)
		if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
			return fmt.Errorf("upsert user %s: %w", seed.Email, err)
		}
	}

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err == nil {
		logger.Info("Users seeded", zap.Int64("count", count))
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
