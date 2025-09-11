package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// TransactionManager provides transaction support for MongoDB operations
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(sessionContext mongo.SessionContext) error) error
}

type transactionManager struct {
	client *mongo.Client
}

func NewTransactionManager(db *MongoDB) TransactionManager {
	return &transactionManager{
		client: db.Client,
	}
}

// WithTransaction executes the provided function within a MongoDB transaction
func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(sessionContext mongo.SessionContext) error) error {
	session, err := tm.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Execute the function within a transaction
	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		_, err := session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			return nil, fn(sc)
		})
		return err
	})
}