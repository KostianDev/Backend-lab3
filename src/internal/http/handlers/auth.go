package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/http/requests"
	"bckndlab3/src/internal/http/responses"
	"bckndlab3/src/internal/storage"
)

// AuthHandler manages authentication related endpoints.
type AuthHandler struct {
	AuthService *storage.AuthService
}

func NewAuthHandler(service *storage.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: service}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
}

// Register creates a new user.
func (h *AuthHandler) Register(c *gin.Context) {
	var req requests.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(responses.NewValidationError(err))
		return
	}

	user, err := h.AuthService.RegisterUser(c.Request.Context(), req.Email, req.Password, req.DefaultCurrency)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, responses.NewUserResponse(user))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(responses.NewValidationError(err))
		return
	}

	user, err := h.AuthService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.NewUserResponse(user))
}
