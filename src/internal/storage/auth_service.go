package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"gorm.io/gorm"

	"bckndlab3/src/internal/models"
)

// AuthService handles user creation and credential verification.
type AuthService struct {
	db       *gorm.DB
	users    *UserRepository
	accounts *AccountRepository
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db:       db,
		users:    NewUserRepository(db),
		accounts: NewAccountRepository(db),
	}
}

// RegisterUser creates a user, hashes the password, and ensures an account exists.
func (s *AuthService) RegisterUser(ctx context.Context, email, password, defaultCurrency string) (*models.User, error) {
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
		account := &models.Account{
			UserID:          user.ID,
			CurrencyISOCode: user.DefaultCurrency,
		}
		if err := tx.WithContext(ctx).Create(account).Error; err != nil {
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
		return nil, ErrPreconditionFailed
	}
	return user, nil
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
