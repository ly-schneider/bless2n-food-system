package helpers

import (
	"context"
	"fmt"
	"time"

	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestDB provides database helper methods for testing
type TestDB struct {
	client   *mongo.Client
	database *mongo.Database
	dbName   string
}

// NewTestDB creates a new test database helper
func NewTestDB(mongoURI, dbName string) (*TestDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)

	return &TestDB{
		client:   client,
		database: database,
		dbName:   dbName,
	}, nil
}

// Close closes the database connection
func (db *TestDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.client.Disconnect(ctx)
}

// CleanAll drops all collections in the test database
func (db *TestDB) CleanAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collections, err := db.database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	for _, collection := range collections {
		if err := db.database.Collection(collection).Drop(ctx); err != nil {
			return fmt.Errorf("failed to drop collection %s: %w", collection, err)
		}
	}

	return nil
}

// SeedUser inserts a test user into the database
func (db *TestDB) SeedUser(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.database.Collection("users")
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to seed user: %w", err)
	}

	return nil
}

// SeedUsers inserts multiple test users into the database
func (db *TestDB) SeedUsers(users []*domain.User) error {
	if len(users) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.database.Collection("users")
	
	// Convert to []interface{} for InsertMany
	docs := make([]interface{}, len(users))
	for i, user := range users {
		docs[i] = user
	}

	_, err := collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (db *TestDB) GetUser(userID string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.database.Collection("users")
	var user domain.User

	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (db *TestDB) GetUserByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.database.Collection("users")
	var user domain.User

	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// CountUsers returns the number of users in the database
func (db *TestDB) CountUsers() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.database.Collection("users")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// CleanOTPs removes all OTP records (for cleanup)
func (db *TestDB) CleanOTPs() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.database.Collection("otp_tokens")
	_, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to clean OTP tokens: %w", err)
	}

	return nil
}