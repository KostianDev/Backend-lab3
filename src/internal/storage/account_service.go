package storage

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bckndlab3/src/internal/models"
)

// AccountService encapsulates transactional account operations.
type AccountService struct {
	db         *gorm.DB
	accounts   *AccountRepository
	users      *UserRepository
	allowDebit bool
}

func NewAccountService(db *gorm.DB, allowDebit bool) *AccountService {
	return &AccountService{
		db:         db,
		accounts:   NewAccountRepository(db),
		users:      NewUserRepository(db),
		allowDebit: allowDebit,
	}
}

// CreditIncome credits an income amount to a user's account.
func (s *AccountService) CreditIncome(ctx context.Context, userID uint, income *models.Income) (*models.Income, int64, error) {
	income.UserID = userID

	var updatedBalance int64

	err := WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		account, err := s.accounts.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		income.AccountID = account.ID

		if err := s.accounts.CreateIncome(ctx, tx, income); err != nil {
			return err
		}

		updatedBalance, err = s.accounts.AdjustBalance(ctx, tx, account.ID, income.AmountCents)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return income, updatedBalance, nil
}

// DebitExpense debits an expense amount from a user's account, respecting overdraft policy.
func (s *AccountService) DebitExpense(ctx context.Context, userID uint, expense *models.Expense) (*models.Expense, int64, error) {
	expense.UserID = userID
	var updatedBalance int64

	err := WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		account, err := s.accounts.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		expense.AccountID = account.ID

		delta := -expense.AmountCents

		newBalance, err := s.accounts.AdjustBalance(ctx, tx, account.ID, delta)
		if err != nil {
			return err
		}

		if !s.allowDebit && newBalance < 0 {
			return ErrPreconditionFailed
		}

		if err := s.accounts.CreateExpense(ctx, tx, expense); err != nil {
			return err
		}

		updatedBalance = newBalance
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return expense, updatedBalance, nil
}

// SetDefaultCurrency updates a user's default currency.
func (s *AccountService) SetDefaultCurrency(ctx context.Context, userID uint, currency string) error {
	err := s.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("default_currency", currency).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

// EnsureAccount ensures an account exists for the given user.
func (s *AccountService) EnsureAccount(ctx context.Context, userID uint, currency string) (*models.Account, error) {
	account, err := s.accounts.GetByUserID(ctx, userID)
	if err == nil {
		return account, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	account = &models.Account{
		UserID:          userID,
		CurrencyISOCode: currency,
	}

	if err := s.db.WithContext(ctx).Create(account).Error; err != nil {
		return nil, translateError(err)
	}
	return account, nil
}
