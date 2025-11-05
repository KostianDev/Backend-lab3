package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"bckndlab3/src/internal/models"
)

// AuthService handles user creation and credential verification.
type AuthService struct {
	db    *gorm.DB
	users *UserRepository
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db:    db,
		users: NewUserRepository(db),
	}
}

// RegisterUser creates a user, hashes the password, and ensures an account exists.
func (s *AuthService) RegisterUser(ctx context.Context, email, password, defaultCurrency string) (*models.User, error) {
	if defaultCurrency == "" {
		defaultCurrency = "UAH"
	} else {
		defaultCurrency = strings.ToUpper(defaultCurrency)
	}

	hashed := hashPassword(password)
	user := &models.User{
		Email:           email,
		PasswordHash:    hashed,
		DefaultCurrency: defaultCurrency,
	}

	err := WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(user).Error; err != nil {
			return translateError(err)
		}
		if err := tx.WithContext(ctx).Create(&models.Account{
			UserID:          user.ID,
			CurrencyISOCode: user.DefaultCurrency,
		}).Error; err != nil {
			return translateError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate validates user credentials.
func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	expected := hashPassword(password)
	if user.PasswordHash != expected {
		return nil, fmt.Errorf("%w: invalid credentials", ErrPreconditionFailed)
	}
	return user, nil
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
