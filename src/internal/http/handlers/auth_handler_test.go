package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"bckndlab3/src/internal/storage"
)

func TestAuthHandlerDeleteUser(t *testing.T) {
	_, authService, _, engine, _ := setupHandlerTest(t)

	ctx := context.Background()
	user, err := authService.RegisterUser(ctx, "handler-delete@example.com", "strongpass", "usd")
	require.NoError(t, err)

	body := strings.NewReader(fmt.Sprintf(`{"email":"%s","password":"strongpass"}`, user.Email))
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/auth/%d", user.ID), body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	_, err = authService.Authenticate(ctx, "handler-delete@example.com", "strongpass")
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestAuthHandlerDeleteUserNotFound(t *testing.T) {
	_, _, _, engine, _ := setupHandlerTest(t)

	body := strings.NewReader(`{"email":"missing@example.com","password":"strongpass"}`)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/999", body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestAuthHandlerDeleteUserForbidden(t *testing.T) {
	_, authService, _, engine, _ := setupHandlerTest(t)

	ctx := context.Background()
	owner, err := authService.RegisterUser(ctx, "owner@example.com", "strongpass", "usd")
	require.NoError(t, err)
	_, err = authService.RegisterUser(ctx, "other@example.com", "strongpass", "usd")
	require.NoError(t, err)

	body := strings.NewReader(`{"email":"other@example.com","password":"strongpass"}`)
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/auth/%d", owner.ID), body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusForbidden, res.Code)

	_, err = authService.Authenticate(ctx, "owner@example.com", "strongpass")
	require.NoError(t, err)
}
