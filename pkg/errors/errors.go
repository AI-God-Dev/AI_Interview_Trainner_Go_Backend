package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrInvalidInput     = NewAppError(http.StatusBadRequest, "Invalid input", nil)
	ErrUnauthorized     = NewAppError(http.StatusUnauthorized, "Unauthorized", nil)
	ErrForbidden        = NewAppError(http.StatusForbidden, "Forbidden", nil)
	ErrNotFound         = NewAppError(http.StatusNotFound, "Resource not found", nil)
	ErrInternalServer   = NewAppError(http.StatusInternalServerError, "Internal server error", nil)
	ErrPaymentRequired  = NewAppError(http.StatusPaymentRequired, "Payment required", nil)
	ErrTooManyRequests  = NewAppError(http.StatusTooManyRequests, "Too many requests", nil)
	ErrServiceUnavailable = NewAppError(http.StatusServiceUnavailable, "Service unavailable", nil)
)

func Wrap(err error, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewAppError(http.StatusInternalServerError, message, err)
}

