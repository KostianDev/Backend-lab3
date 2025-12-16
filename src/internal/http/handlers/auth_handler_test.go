package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"bckndlab3/src/internal/storage"
)

func TestAuthHandlerDeleteUser(t *testing.T) {
	env := setupHandlerTest(t)

	ctx := context.Background()
	user, err := env.authService.RegisterUser(ctx, "handler-delete@example.com", "strongpass", "usd")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", env.authHeader(user.ID, user.Email))
	res := httptest.NewRecorder()

	env.engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	_, err = env.authService.Authenticate(ctx, "handler-delete@example.com", "strongpass")
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestAuthHandlerDeleteUserUnauthorized(t *testing.T) {
	env := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/me", nil)
	res := httptest.NewRecorder()

	env.engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}
