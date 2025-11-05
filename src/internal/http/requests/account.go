package requests

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/models"
)

// IncomeRequest represents payload for creating an income record.
type IncomeRequest struct {
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Source     string  `json:"source" binding:"required"`
	ReceivedAt string  `json:"received_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Notes      string  `json:"notes" binding:"omitempty,max=512"`
}

// ToModel converts request to models.Income.
func (r IncomeRequest) ToModel(defaultTime time.Time) *models.Income {
	ts := defaultTime
	if r.ReceivedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, r.ReceivedAt); err == nil {
			ts = parsed
		}
	}

	return &models.Income{
		AmountCents: int64(math.Round(r.Amount * 100)),
		Source:      r.Source,
		ReceivedAt:  ts,
		Notes:       r.Notes,
	}
}

// ExpenseRequest represents payload for creating an expense record.
type ExpenseRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required"`
	IncurredAt  string  `json:"incurred_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Description string  `json:"description" binding:"omitempty,max=512"`
}

// ToModel converts request to models.Expense.
func (r ExpenseRequest) ToModel(defaultTime time.Time) *models.Expense {
	ts := defaultTime
	if r.IncurredAt != "" {
		if parsed, err := time.Parse(time.RFC3339, r.IncurredAt); err == nil {
			ts = parsed
		}
	}

	return &models.Expense{
		AmountCents: int64(math.Round(r.Amount * 100)),
		Category:    r.Category,
		IncurredAt:  ts,
		Description: r.Description,
	}
}

// ParseUintParam parses uint from path parameter.
func ParseUintParam(c *gin.Context, key string) (uint, error) {
	value := c.Param(key)
	if value == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return uint(parsed), nil
}

// ParseLimitQuery parses an optional limit query parameter.
func ParseLimitQuery(c *gin.Context, key string, fallback int) int {
	value := c.DefaultQuery(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
