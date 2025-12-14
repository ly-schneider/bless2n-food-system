package seed

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Seeder defines a unit of seed data for a single collection.
type Seeder interface {
	Name() string
	Seed(ctx context.Context, db *mongo.Database) error
}

// RunAll executes each seeder in order and wraps any failure with the seeder name.
func RunAll(ctx context.Context, db *mongo.Database, seeders []Seeder) error {
	for _, seeder := range seeders {
		if seeder == nil {
			continue
		}
		if err := seeder.Seed(ctx, db); err != nil {
			return fmt.Errorf("%s seeder failed: %w", seeder.Name(), err)
		}
	}
	return nil
}
