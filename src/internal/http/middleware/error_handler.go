package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/http/responses"
	"bckndlab3/src/internal/storage"
)

// ErrorHandler converts domain errors into JSON HTTP responses.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status, payload := mapError(err)

		c.AbortWithStatusJSON(status, payload)
	}
}

func mapError(err error) (int, gin.H) {
	var (
		status = http.StatusInternalServerError
		code   = "internal_error"
	)

	if ve, ok := responses.ExtractValidationError(err); ok {
		return http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "validation_error",
				"message": ve.Error(),
			},
		}
	}

	switch {
	case errors.Is(err, storage.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, storage.ErrConflict):
		status = http.StatusConflict
		code = "conflict"
	case errors.Is(err, storage.ErrInsufficientFunds):
		status = http.StatusBadRequest
		code = "insufficient_funds"
	case errors.Is(err, storage.ErrPreconditionFailed):
		status = http.StatusBadRequest
		code = "precondition_failed"
	}

	return status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": err.Error(),
		},
	}
}
