package responses

import (
	"time"

	"bckndlab3/src/internal/models"
)

// IncomeResponse payload for created income that returns current balance context.
type IncomeResponse struct {
	ID           uint    `json:"id"`
	Amount       float64 `json:"amount"`
	Source       string  `json:"source"`
	ReceivedAt   string  `json:"received_at"`
	Notes        string  `json:"notes,omitempty"`
	BalanceCents int64   `json:"balance_cents"`
}

// NewIncomeResponse builds an IncomeResponse.
func NewIncomeResponse(income *models.Income, balance int64) IncomeResponse {
	return IncomeResponse{
		ID:           income.ID,
		Amount:       centsToFloat(income.AmountCents),
		Source:       income.Source,
		ReceivedAt:   income.ReceivedAt.Format(time.RFC3339),
		Notes:        income.Notes,
		BalanceCents: balance,
	}
}

// IncomeListItem represents income data without balance context.
type IncomeListItem struct {
	ID         uint    `json:"id"`
	Amount     float64 `json:"amount"`
	Source     string  `json:"source"`
	ReceivedAt string  `json:"received_at"`
	Notes      string  `json:"notes,omitempty"`
}

// NewIncomeListResponse builds a list of incomes for listing endpoints.
func NewIncomeListResponse(incomes []models.Income) []IncomeListItem {
	items := make([]IncomeListItem, 0, len(incomes))
	for i := range incomes {
		items = append(items, IncomeListItem{
			ID:         incomes[i].ID,
			Amount:     centsToFloat(incomes[i].AmountCents),
			Source:     incomes[i].Source,
			ReceivedAt: incomes[i].ReceivedAt.Format(time.RFC3339),
			Notes:      incomes[i].Notes,
		})
	}
	return items
}

// ExpenseResponse payload for created expense.
type ExpenseResponse struct {
	ID           uint    `json:"id"`
	Amount       float64 `json:"amount"`
	Category     string  `json:"category"`
	IncurredAt   string  `json:"incurred_at"`
	Description  string  `json:"description,omitempty"`
	BalanceCents int64   `json:"balance_cents"`
}

// NewExpenseResponse builds an ExpenseResponse.
func NewExpenseResponse(expense *models.Expense, balance int64) ExpenseResponse {
	return ExpenseResponse{
		ID:           expense.ID,
		Amount:       centsToFloat(expense.AmountCents),
		Category:     expense.Category,
		IncurredAt:   expense.IncurredAt.Format(time.RFC3339),
		Description:  expense.Description,
		BalanceCents: balance,
	}
}

// ExpenseListItem represents expense data without balance context.
type ExpenseListItem struct {
	ID          uint    `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	IncurredAt  string  `json:"incurred_at"`
	Description string  `json:"description,omitempty"`
}

// NewExpenseListResponse builds a list of expenses for listing endpoints.
func NewExpenseListResponse(expenses []models.Expense) []ExpenseListItem {
	items := make([]ExpenseListItem, 0, len(expenses))
	for i := range expenses {
		items = append(items, ExpenseListItem{
			ID:          expenses[i].ID,
			Amount:      centsToFloat(expenses[i].AmountCents),
			Category:    expenses[i].Category,
			IncurredAt:  expenses[i].IncurredAt.Format(time.RFC3339),
			Description: expenses[i].Description,
		})
	}
	return items
}

// BalanceResponse returns account balance detail.
type BalanceResponse struct {
	AccountID       uint   `json:"account_id"`
	BalanceCents    int64  `json:"balance_cents"`
	CurrencyISOCode string `json:"currency_iso_code"`
}

// NewBalanceResponse builds a balance response payload.
func NewBalanceResponse(account *models.Account) BalanceResponse {
	return BalanceResponse{
		AccountID:       account.ID,
		BalanceCents:    account.BalanceCents,
		CurrencyISOCode: account.CurrencyISOCode,
	}
}

func centsToFloat(cents int64) float64 {
	return float64(cents) / 100
}
