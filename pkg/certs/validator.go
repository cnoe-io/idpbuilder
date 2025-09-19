package certs

import "fmt"

// ValidationMode defines the validation strictness level
type ValidationMode int

const (
	// StrictMode performs full certificate validation including chain verification
	StrictMode ValidationMode = iota

	// LenientMode performs basic certificate validation with relaxed chain requirements
	LenientMode

	// InsecureMode performs minimal validation (for development/testing only)
	InsecureMode
)

// String returns the string representation of ValidationMode
func (vm ValidationMode) String() string {
	switch vm {
	case StrictMode:
		return "Strict"
	case LenientMode:
		return "Lenient"
	case InsecureMode:
		return "Insecure"
	default:
		return fmt.Sprintf("Unknown(%d)", int(vm))
	}
}