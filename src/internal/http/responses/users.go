package responses

import "bckndlab3/src/internal/models"

// UserResponse represents a user payload returned by the API.
type UserResponse struct {
	ID              uint   `json:"id"`
	Email           string `json:"email"`
	DefaultCurrency string `json:"default_currency"`
}

// NewUserResponse builds a response from the given user model.
func NewUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		DefaultCurrency: user.DefaultCurrency,
	}
}
