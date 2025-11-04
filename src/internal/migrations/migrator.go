package migrations

import (
	"fmt"

	"gorm.io/gorm"

	"bckndlab3/src/internal/models"
)

// Run ensures all database schema migrations are applied.
func Run(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Income{},
		&models.Expense{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}
