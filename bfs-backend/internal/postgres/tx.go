package postgres

import (
	"context"

	"gorm.io/gorm"
)

// TxManager manages database transactions.
type TxManager struct {
	db *gorm.DB
}

// NewTxManager creates a new transaction manager.
func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes the given function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (m *TxManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return m.db.WithContext(ctx).Transaction(fn)
}

// DB returns the underlying GORM DB instance.
func (m *TxManager) DB() *gorm.DB {
	return m.db
}
