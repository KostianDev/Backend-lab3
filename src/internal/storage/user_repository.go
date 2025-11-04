package storage

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"bckndlab3/src/internal/models"
)

// UserRepository exposes persistence operations for user entities.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create persists a new user and related account.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return translateError(err)
	}
	return nil
}

// GetByEmail fetches a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, translateError(err)
	}
	return &user, nil
}

// GetByID fetches a user by primary key.
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, translateError(err)
	}
	return &user, nil
}

func translateError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	return fmt.Errorf("storage: %w", err)
}
