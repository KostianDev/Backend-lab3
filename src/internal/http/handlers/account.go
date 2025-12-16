package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/http/middleware"
	"bckndlab3/src/internal/http/requests"
	"bckndlab3/src/internal/http/responses"
	"bckndlab3/src/internal/services"
	"bckndlab3/src/internal/storage"
)

// AccountHandler manages account, income, and expense endpoints.
type AccountHandler struct {
	Service *storage.AccountService
	Time    services.TimeProvider
}

func NewAccountHandler(service *storage.AccountService, timeProvider services.TimeProvider) *AccountHandler {
	return &AccountHandler{Service: service, Time: timeProvider}
}

func (h *AccountHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/incomes", h.CreateIncome)
	router.POST("/expenses", h.CreateExpense)
	router.GET("/balance", h.GetBalance)
	router.GET("/incomes", h.ListIncomes)
	router.GET("/expenses", h.ListExpenses)
}

func (h *AccountHandler) CreateIncome(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "unauthorized", "message": "user not authenticated"},
		})
		return
	}

	var req requests.IncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(responses.NewValidationError(err))
		return
	}

	incomeModel := req.ToModel(h.Time.Now())
	income, balance, err := h.Service.CreditIncome(c.Request.Context(), userID, incomeModel)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, responses.NewIncomeResponse(income, balance))
}

func (h *AccountHandler) CreateExpense(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "unauthorized", "message": "user not authenticated"},
		})
		return
	}

	var req requests.ExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(responses.NewValidationError(err))
		return
	}

	expenseModel := req.ToModel(h.Time.Now())
	expense, balance, err := h.Service.DebitExpense(c.Request.Context(), userID, expenseModel)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, responses.NewExpenseResponse(expense, balance))
}

func (h *AccountHandler) GetBalance(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "unauthorized", "message": "user not authenticated"},
		})
		return
	}

	account, err := h.Service.GetAccountByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.NewBalanceResponse(account))
}

func (h *AccountHandler) ListIncomes(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "unauthorized", "message": "user not authenticated"},
		})
		return
	}
	limit := requests.ParseLimitQuery(c, "limit", 50)

	account, err := h.Service.GetAccountByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	incomes, err := h.Service.ListIncomes(c.Request.Context(), account.ID, limit)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.NewIncomeListResponse(incomes))
}

func (h *AccountHandler) ListExpenses(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "unauthorized", "message": "user not authenticated"},
		})
		return
	}
	limit := requests.ParseLimitQuery(c, "limit", 50)

	account, err := h.Service.GetAccountByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	expenses, err := h.Service.ListExpenses(c.Request.Context(), account.ID, limit)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.NewExpenseListResponse(expenses))
}
