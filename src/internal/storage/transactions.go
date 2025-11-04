package storage

import (
	"context"

	"gorm.io/gorm"
)

// WithTransaction executes the supplied callback within a database transaction.
func WithTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}
