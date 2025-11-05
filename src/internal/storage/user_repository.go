package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	sqlite3 "github.com/mattn/go-sqlite3"
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

// DeleteByID removes a user and cascades related aggregates.
func (r *UserRepository) DeleteByID(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if err := result.Error; err != nil {
		return translateError(err)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func translateError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrConflict
		case "23503":
			return ErrPreconditionFailed
		}
	}
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.ExtendedCode {
		case sqlite3.ErrConstraintUnique, sqlite3.ErrConstraintPrimaryKey:
			return ErrConflict
		}
	}
	return fmt.Errorf("storage: %w", err)
}
