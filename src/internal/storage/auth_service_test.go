package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceRegisterAndAuthenticate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	service := NewAuthService(db)

	user, err := service.RegisterUser(ctx, "auth@example.com", "topsecret", "gbp")
	require.NoError(t, err)
	require.Equal(t, "GBP", user.DefaultCurrency)
	require.NotEqual(t, "topsecret", user.PasswordHash)

	accountRepo := NewAccountRepository(db)
	account, err := accountRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "GBP", account.CurrencyISOCode)

	lookedUp, err := service.Authenticate(ctx, "auth@example.com", "topsecret")
	require.NoError(t, err)
	require.Equal(t, user.ID, lookedUp.ID)

	_, err = service.Authenticate(ctx, "auth@example.com", "wrongsecret")
	require.ErrorIs(t, err, ErrPreconditionFailed)
}

func TestAuthServiceDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	service := NewAuthService(db)

	user, err := service.RegisterUser(ctx, "delete@example.com", "strongpass", "usd")
	require.NoError(t, err)

	// sanity check account exists
	accountRepo := NewAccountRepository(db)
	_, err = accountRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)

	require.NoError(t, service.DeleteUser(ctx, user.ID))

	_, err = service.Authenticate(ctx, "delete@example.com", "strongpass")
	require.ErrorIs(t, err, ErrNotFound)

	err = service.DeleteUser(ctx, user.ID)
	require.ErrorIs(t, err, ErrNotFound)

	_, err = accountRepo.GetByUserID(ctx, user.ID)
	require.ErrorIs(t, err, ErrNotFound)
}
