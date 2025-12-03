package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AppError
		want    string
		contains string
	}{
		{
			name: "error with message only",
			err:  NewAppError(http.StatusBadRequest, "test error", nil),
			contains: "test error",
		},
		{
			name: "error with underlying error",
			err:  NewAppError(http.StatusInternalServerError, "wrapped", errors.New("underlying")),
			contains: "wrapped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got == "" {
				t.Error("Error() should not return empty string")
			}
			if tt.contains != "" && len(got) < len(tt.contains) {
				t.Errorf("Error() = %v, should contain %v", got, tt.contains)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := NewAppError(http.StatusInternalServerError, "wrapped", underlying)

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}

	errNoWrap := NewAppError(http.StatusBadRequest, "no wrap", nil)
	if errNoWrap.Unwrap() != nil {
		t.Error("Unwrap() should return nil when no underlying error")
	}
}

func TestWrap(t *testing.T) {
	underlying := errors.New("underlying")
	wrapped := Wrap(underlying, "wrapped message")

	if wrapped == nil {
		t.Fatal("Wrap() should not return nil")
	}

	if wrapped.Message != "wrapped message" {
		t.Errorf("Wrap() message = %v, want wrapped message", wrapped.Message)
	}

	// Test wrapping an AppError returns the same error
	appErr := NewAppError(http.StatusBadRequest, "app error", nil)
	wrappedAppErr := Wrap(appErr, "new message")
	if wrappedAppErr != appErr {
		t.Error("Wrap() should return the same AppError when wrapping an AppError")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		code int
	}{
		{"ErrInvalidInput", ErrInvalidInput, http.StatusBadRequest},
		{"ErrUnauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"ErrForbidden", ErrForbidden, http.StatusForbidden},
		{"ErrNotFound", ErrNotFound, http.StatusNotFound},
		{"ErrInternalServer", ErrInternalServer, http.StatusInternalServerError},
		{"ErrPaymentRequired", ErrPaymentRequired, http.StatusPaymentRequired},
		{"ErrTooManyRequests", ErrTooManyRequests, http.StatusTooManyRequests},
		{"ErrServiceUnavailable", ErrServiceUnavailable, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("%s should not be nil", tt.name)
			}
			if tt.err.Code != tt.code {
				t.Errorf("%s.Code = %v, want %v", tt.name, tt.err.Code, tt.code)
			}
			if tt.err.Message == "" {
				t.Errorf("%s.Message should not be empty", tt.name)
			}
		})
	}
}

