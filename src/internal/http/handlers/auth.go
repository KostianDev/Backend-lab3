package handlers

import (
	"crypto/rand"
	"encoding/hex"
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
	router.DELETE("/:userID", h.Delete)
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

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": hex.EncodeToString(tokenBytes),
		"user":  responses.NewUserResponse(user),
	})
}

// Delete removes a user by identifier.
func (h *AuthHandler) Delete(c *gin.Context) {
	userID, err := requests.ParseUintParam(c, "userID")
	if err != nil {
		c.Error(responses.NewValidationError(err))
		return
	}

	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(responses.NewValidationError(err))
		return
	}

	authenticated, err := h.AuthService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	if authenticated.ID != userID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "forbidden",
				"message": "cannot delete another user",
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
