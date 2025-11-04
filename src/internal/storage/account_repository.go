package storage

import (
	"context"

	"gorm.io/gorm"

	"bckndlab3/src/internal/models"
)

// AccountRepository handles persistence for account aggregates.
type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// GetByUserID returns the account belonging to a specific user.
func (r *AccountRepository) GetByUserID(ctx context.Context, userID uint) (*models.Account, error) {
	var account models.Account
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		return nil, translateError(err)
	}
	return &account, nil
}

// UpdateBalance sets the account balance to the provided amount.
func (r *AccountRepository) UpdateBalance(ctx context.Context, accountID uint, balanceCents int64) error {
	result := r.db.WithContext(ctx).Model(&models.Account{}).
		Where("id = ?", accountID).
		Update("balance_cents", balanceCents)

	if err := result.Error; err != nil {
		return translateError(err)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// AdjustBalance increments the account balance by delta and returns the updated value.
func (r *AccountRepository) AdjustBalance(ctx context.Context, tx *gorm.DB, accountID uint, delta int64) (int64, error) {
	type balanceRow struct {
		BalanceCents int64
	}

	tx = tx.WithContext(ctx)

	if err := tx.Exec("UPDATE accounts SET balance_cents = balance_cents + ? WHERE id = ?", delta, accountID).Error; err != nil {
		return 0, translateError(err)
	}

	var row balanceRow
	if err := tx.Raw("SELECT balance_cents FROM accounts WHERE id = ?", accountID).Scan(&row).Error; err != nil {
		return 0, translateError(err)
	}

	return row.BalanceCents, nil
}

// CreateIncome records a new income entry tied to the account.
func (r *AccountRepository) CreateIncome(ctx context.Context, tx *gorm.DB, income *models.Income) error {
	if err := tx.WithContext(ctx).Create(income).Error; err != nil {
		return translateError(err)
	}
	return nil
}

// CreateExpense records a new expense entry tied to the account.
func (r *AccountRepository) CreateExpense(ctx context.Context, tx *gorm.DB, expense *models.Expense) error {
	if err := tx.WithContext(ctx).Create(expense).Error; err != nil {
		return translateError(err)
	}
	return nil
}

// ListIncomes retrieves incomes for an account ordered by most recent.
func (r *AccountRepository) ListIncomes(ctx context.Context, accountID uint, limit int) ([]models.Income, error) {
	var incomes []models.Income
	query := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("received_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&incomes).Error; err != nil {
		return nil, translateError(err)
	}
	return incomes, nil
}

// ListExpenses retrieves expenses for an account ordered by most recent.
func (r *AccountRepository) ListExpenses(ctx context.Context, accountID uint, limit int) ([]models.Expense, error) {
	var expenses []models.Expense
	query := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("incurred_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&expenses).Error; err != nil {
		return nil, translateError(err)
	}
	return expenses, nil
}
