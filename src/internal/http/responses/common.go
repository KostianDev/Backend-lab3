package responses

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValidationError wraps request validation issues.
type ValidationError struct {
	Err error
}

func (v ValidationError) Error() string { return v.Err.Error() }

// NewValidationError creates a new validation error instance.
func NewValidationError(err error) error {
	return ValidationError{Err: err}
}

// HandleValidationError converts a validation error to JSON response.
func HandleValidationError(c *gin.Context, err ValidationError) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error": gin.H{
			"code":    "validation_error",
			"message": err.Error(),
		},
	})
}

// ExtractValidationError checks if the provided error is a ValidationError.
func ExtractValidationError(err error) (ValidationError, bool) {
	var ve ValidationError
	if errors.As(err, &ve) {
		return ve, true
	}
	return ValidationError{}, false
}
