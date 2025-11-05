package requests

// RegisterRequest represents the incoming payload for user registration.
type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	DefaultCurrency string `json:"default_currency" binding:"omitempty,len=3"`
}

// LoginRequest represents credentials for authentication.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
