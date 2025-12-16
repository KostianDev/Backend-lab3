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
	"bckndlab3/src/internal/models"
	"bckndlab3/src/internal/storage"
)

type fixedTimeProvider struct {
	value time.Time
}

func (f fixedTimeProvider) Now() time.Time { return f.value }

type testEnv struct {
	db             *gorm.DB
	authService    *storage.AuthService
	accountService *storage.AccountService
	jwtService     *storage.JWTService
	engine         *gin.Engine
	frozen         time.Time
}

func setupHandlerTest(t *testing.T) *testEnv {
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
	jwtService := storage.NewJWTService("test-secret-key", 24*time.Hour)

	frozen := time.Date(2025, time.November, 5, 12, 0, 0, 0, time.UTC)

	engine := router.New(router.Dependencies{
		Auth:       handlers.NewAuthHandler(authService, jwtService),
		Account:    handlers.NewAccountHandler(accountService, fixedTimeProvider{value: frozen}),
		JWTService: jwtService,
	})

	return &testEnv{
		db:             db,
		authService:    authService,
		accountService: accountService,
		jwtService:     jwtService,
		engine:         engine,
		frozen:         frozen,
	}
}

func (e *testEnv) authHeader(userID uint, email string) string {
	token, _ := e.jwtService.GenerateToken(userID, email)
	return "Bearer " + token
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
	env := setupHandlerTest(t)

	ctx := context.Background()
	user, err := env.authService.RegisterUser(ctx, "handler@example.com", "password123", "uah")
	require.NoError(t, err)

	authHeader := env.authHeader(user.ID, user.Email)

	t.Run("create income", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			"amount": 100.5,
			"source": "Salary",
			"notes":  "Project X",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/incomes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		res := httptest.NewRecorder()

		env.engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusCreated, res.Code)

		var payload incomeResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Equal(t, int64(10050), payload.BalanceCents)
		require.Equal(t, env.frozen.Format(time.RFC3339), payload.ReceivedAt)
	})

	t.Run("create expense", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			"amount":   20.0,
			"category": "Meals",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/expenses", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		res := httptest.NewRecorder()

		env.engine.ServeHTTP(res, req)
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

		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/expenses", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		res := httptest.NewRecorder()

		env.engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusBadRequest, res.Code)

		var payload errorEnvelope
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Equal(t, "insufficient_funds", payload.Error.Code)
	})

	t.Run("list incomes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/incomes", nil)
		req.Header.Set("Authorization", authHeader)
		res := httptest.NewRecorder()

		env.engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code)

		var payload []incomeListItem
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Len(t, payload, 1)
		require.InDelta(t, 100.5, payload[0].Amount, 0.001)
	})

	t.Run("balance endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/balance", nil)
		req.Header.Set("Authorization", authHeader)
		res := httptest.NewRecorder()

		env.engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code)

		var payload balanceResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
		require.Equal(t, int64(8050), payload.BalanceCents)
	})

	// ensure database state matches expectations
	account, err := env.accountService.GetAccountByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(8050), account.BalanceCents)
}

func TestAccountHandlerCreateIncomeValidationError(t *testing.T) {
	env := setupHandlerTest(t)

	ctx := context.Background()
	user, err := env.authService.RegisterUser(ctx, "invalid-income@example.com", "password123", "uah")
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{
		"source": "Gift",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/incomes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", env.authHeader(user.ID, user.Email))
	res := httptest.NewRecorder()

	env.engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)

	var payload errorEnvelope
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, "validation_error", payload.Error.Code)
}

func TestAccountHandlerGetBalanceNotFound(t *testing.T) {
	env := setupHandlerTest(t)

	// Generate token for non-existent user
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/balance", nil)
	req.Header.Set("Authorization", env.authHeader(999, "nonexistent@example.com"))
	res := httptest.NewRecorder()

	env.engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)

	var payload errorEnvelope
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, "not_found", payload.Error.Code)
}

func TestAccountHandlerListIncomesRespectLimit(t *testing.T) {
	env := setupHandlerTest(t)

	ctx := context.Background()
	user, err := env.authService.RegisterUser(ctx, "limit@example.com", "password123", "uah")
	require.NoError(t, err)

	amounts := []int64{10000, 20000, 30000}
	for i, cents := range amounts {
		income := &models.Income{
			AmountCents: cents,
			Source:      fmt.Sprintf("src-%d", i),
			ReceivedAt:  env.frozen.Add(time.Duration(i) * time.Hour),
		}
		_, _, err := env.accountService.CreditIncome(ctx, user.ID, income)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/incomes?limit=2", nil)
	req.Header.Set("Authorization", env.authHeader(user.ID, user.Email))
	res := httptest.NewRecorder()

	env.engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	var payload []struct {
		Amount float64 `json:"amount"`
	}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Len(t, payload, 2)
	require.InDelta(t, 300.0, payload[0].Amount, 0.001)
	require.InDelta(t, 200.0, payload[1].Amount, 0.001)
}

func TestAccountHandlerListExpensesRespectLimit(t *testing.T) {
	env := setupHandlerTest(t)

	ctx := context.Background()
	user, err := env.authService.RegisterUser(ctx, "limit-expenses@example.com", "password123", "uah")
	require.NoError(t, err)

	_, _, err = env.accountService.CreditIncome(ctx, user.ID, &models.Income{AmountCents: 100000, Source: "seed", ReceivedAt: env.frozen})
	require.NoError(t, err)

	costs := []int64{1000, 2000, 3000}
	for i, cents := range costs {
		expense := &models.Expense{
			AmountCents: cents,
			Category:    fmt.Sprintf("cat-%d", i),
			IncurredAt:  env.frozen.Add(time.Duration(i) * time.Hour),
		}
		_, _, err := env.accountService.DebitExpense(ctx, user.ID, expense)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/expenses?limit=2", nil)
	req.Header.Set("Authorization", env.authHeader(user.ID, user.Email))
	res := httptest.NewRecorder()

	env.engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	var payload []struct {
		Amount float64 `json:"amount"`
	}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Len(t, payload, 2)
	require.InDelta(t, 30.0, payload[0].Amount, 0.001)
	require.InDelta(t, 20.0, payload[1].Amount, 0.001)
}
