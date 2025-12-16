package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/http/middleware"
	"bckndlab3/src/internal/http/requests"
	"bckndlab3/src/internal/http/responses"
	"bckndlab3/src/internal/storage"
)

// AuthHandler manages authentication related endpoints.
type AuthHandler struct {
	AuthService *storage.AuthService
	JWTService  *storage.JWTService
}

func NewAuthHandler(authService *storage.AuthService, jwtService *storage.JWTService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
		JWTService:  jwtService,
	}
}

// RegisterPublicRoutes sets up routes that don't require authentication.
func (h *AuthHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
}

// RegisterProtectedRoutes sets up routes that require authentication.
func (h *AuthHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	router.DELETE("/me", h.Delete)
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

	token, err := h.JWTService.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  responses.NewUserResponse(user),
	})
}

// Delete removes the authenticated user.
func (h *AuthHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "unauthorized",
				"message": "user not authenticated",
			},
		})
		return
	}

	if err := h.AuthService.DeleteUser(c.Request.Context(), userID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
