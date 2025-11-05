package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bckndlab3/src/internal/migrations"
	"bckndlab3/src/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, migrations.Run(db))
	require.NoError(t, db.Exec("PRAGMA foreign_keys = ON;").Error)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func TestAccountServiceCreditIncome(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	auth := NewAuthService(db)
	user, err := auth.RegisterUser(ctx, "credit@example.com", "strongpass", "uah")
	require.NoError(t, err)

	svc := NewAccountService(db, false)

	income := &models.Income{AmountCents: 12500, Source: "Salary", Notes: "October"}
	saved, balance, err := svc.CreditIncome(ctx, user.ID, income)
	require.NoError(t, err)
	require.NotZero(t, saved.ID)
	require.False(t, saved.ReceivedAt.IsZero())
	require.Equal(t, int64(12500), balance)

	account, err := svc.GetAccountByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(12500), account.BalanceCents)
}

func TestAccountServiceDebitExpenseInsufficientFunds(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	auth := NewAuthService(db)
	user, err := auth.RegisterUser(ctx, "debit@example.com", "strongpass", "usd")
	require.NoError(t, err)

	svc := NewAccountService(db, false)

	expense := &models.Expense{AmountCents: 5000, Category: "Groceries"}
	_, _, err = svc.DebitExpense(ctx, user.ID, expense)
	require.ErrorIs(t, err, ErrInsufficientFunds)

	account, err := svc.GetAccountByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), account.BalanceCents)
}

func TestAccountServiceDebitExpenseAllowsNegativeBalance(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	auth := NewAuthService(db)
	user, err := auth.RegisterUser(ctx, "allow@example.com", "strongpass", "eur")
	require.NoError(t, err)

	svc := NewAccountService(db, true)

	_, balance, err := svc.CreditIncome(ctx, user.ID, &models.Income{
		AmountCents: 10000,
		Source:      "Adhoc",
		ReceivedAt:  time.Now().UTC(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(10000), balance)

	expense := &models.Expense{AmountCents: 15000, Category: "Equipment"}
	_, balance, err = svc.DebitExpense(ctx, user.ID, expense)
	require.NoError(t, err)
	require.Equal(t, int64(-5000), balance)
}

func TestAccountServiceCreditIncomeRejectsNonPositiveAmount(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	auth := NewAuthService(db)
	user, err := auth.RegisterUser(ctx, "reject-income@example.com", "strongpass", "usd")
	require.NoError(t, err)

	svc := NewAccountService(db, false)

	_, _, err = svc.CreditIncome(ctx, user.ID, &models.Income{AmountCents: 0, Source: "Gift"})
	require.ErrorIs(t, err, ErrPreconditionFailed)

	account, err := svc.GetAccountByUserID(ctx, user.ID)
	require.NoError(t, err)

	incomes, err := svc.ListIncomes(ctx, account.ID, 10)
	require.NoError(t, err)
	require.Empty(t, incomes)
}

func TestAccountServiceDebitExpenseRejectsNonPositiveAmount(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	auth := NewAuthService(db)
	user, err := auth.RegisterUser(ctx, "reject-expense@example.com", "strongpass", "usd")
	require.NoError(t, err)

	svc := NewAccountService(db, false)

	_, _, err = svc.DebitExpense(ctx, user.ID, &models.Expense{AmountCents: 0, Category: "Misc"})
	require.ErrorIs(t, err, ErrPreconditionFailed)

	account, err := svc.GetAccountByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), account.BalanceCents)

	expenses, err := svc.ListExpenses(ctx, account.ID, 10)
	require.NoError(t, err)
	require.Empty(t, expenses)
}

func TestAccountServiceSetDefaultCurrencyUpdatesAccount(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	auth := NewAuthService(db)
	user, err := auth.RegisterUser(ctx, "currency@example.com", "strongpass", "usd")
	require.NoError(t, err)

	svc := NewAccountService(db, false)

	require.NoError(t, svc.SetDefaultCurrency(ctx, user.ID, "pln"))

	var refreshedUser models.User
	require.NoError(t, db.WithContext(ctx).First(&refreshedUser, user.ID).Error)
	require.Equal(t, "PLN", refreshedUser.DefaultCurrency)

	account, err := svc.GetAccountByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "PLN", account.CurrencyISOCode)
}
