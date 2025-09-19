package types

import "fmt"

// RegistryError represents a registry-related error
type RegistryError struct {
	Code    string
	Message string
	Detail  interface{}
}

func (e *RegistryError) Error() string {
	if e.Detail != nil {
		return fmt.Sprintf("%s: %s (detail: %v)", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeConnectionFailed = "CONNECTION_FAILED"
	ErrCodeInvalidConfig    = "INVALID_CONFIG"
	ErrCodeTLSVerification  = "TLS_VERIFICATION_FAILED"
)

// Error constructors
func NewUnauthorizedError(msg string) error {
	return &RegistryError{Code: ErrCodeUnauthorized, Message: msg}
}

func NewNotFoundError(msg string) error {
	return &RegistryError{Code: ErrCodeNotFound, Message: msg}
}

func NewConnectionError(msg string, detail interface{}) error {
	return &RegistryError{Code: ErrCodeConnectionFailed, Message: msg, Detail: detail}
}