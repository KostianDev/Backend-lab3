package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	handlers "bckndlab3/src/internal/http/handlers"
	"bckndlab3/src/internal/http/router"
	"bckndlab3/src/internal/migrations"
	"bckndlab3/src/internal/storage"
)

type fixedTimeProvider struct {
	value time.Time
}

func (f fixedTimeProvider) Now() time.Time { return f.value }

func setupHandlerTest(t *testing.T) (*gorm.DB, *storage.AuthService, *storage.AccountService, *gin.Engine, time.Time) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, migrations.Run(db))
	require.NoError(t, db.Exec("PRAGMA foreign_keys = ON;").Error)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	authService := storage.NewAuthService(db)
	accountService := storage.NewAccountService(db, false)

	frozen := time.Date(2025, time.November, 5, 12, 0, 0, 0, time.UTC)

	engine := router.New(router.Dependencies{
		Auth:    handlers.NewAuthHandler(authService),
		Account: handlers.NewAccountHandler(accountService, fixedTimeProvider{value: frozen}),
	})

	return db, authService, accountService, engine, frozen
}

type incomeResponse struct {
	BalanceCents int64  `json:"balance_cents"`
	ReceivedAt   string `json:"received_at"`
}

type expenseResponse struct {
	BalanceCents int64 `json:"balance_cents"`
}

type errorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

type incomeListItem struct {
	Amount float64 `json:"amount"`
}

type balanceResponse struct {
	BalanceCents int64 `json:"balance_cents"`
}

func TestAccountHandlerIncomeExpenseFlow(t *testing.T) {
	_, authService, accountService, engine, frozen := setupHandlerTest(t)

	ctx := context.Background()
	user, err := authService.RegisterUser(ctx, "handler@example.com", "password123", "uah")
	require.NoError(t, err)

	t.Run("create income", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			"amount": 100.5,
			"source": "Salary",
			"notes":  "Project X",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/accounts/%d/incomes", user.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusCreated, res.Code)

		var payload incomeResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Equal(t, int64(10050), payload.BalanceCents)
		require.Equal(t, frozen.Format(time.RFC3339), payload.ReceivedAt)
	})

	t.Run("create expense", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			"amount":   20.0,
			"category": "Meals",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/accounts/%d/expenses", user.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusCreated, res.Code)

		var payload expenseResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Equal(t, int64(8050), payload.BalanceCents)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			"amount":   1000.0,
			"category": "Equipment",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/accounts/%d/expenses", user.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusBadRequest, res.Code)

		var payload errorEnvelope
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Equal(t, "insufficient_funds", payload.Error.Code)
	})

	t.Run("list incomes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/accounts/%d/incomes", user.ID), nil)
		res := httptest.NewRecorder()

		engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code)

		var payload []incomeListItem
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Len(t, payload, 1)
		require.InDelta(t, 100.5, payload[0].Amount, 0.001)
	})

	t.Run("balance endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/accounts/%d/balance", user.ID), nil)
		res := httptest.NewRecorder()

		engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code)

		var payload balanceResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Equal(t, int64(8050), payload.BalanceCents)
	})

	// ensure database state matches expectations
	account, err := accountService.GetAccountByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(8050), account.BalanceCents)
}
