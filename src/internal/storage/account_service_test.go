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
