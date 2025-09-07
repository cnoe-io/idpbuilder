package certs

import (
	"time"
)

// CertDiagnostics contains diagnostic information about certificate validation
type CertDiagnostics struct {
	// Basic certificate information
	Subject          string
	Issuer           string  
	NotBefore        time.Time
	NotAfter         time.Time
	
	// Chain information
	ChainLength      int
	
	// Validation results
	ValidationErrors []string
	IsExpired        bool
	IsSelfSigned     bool
	
	// Timing information
	ValidationStarted   time.Time
	ValidationCompleted time.Time
	
	// General message for cases where no validation was performed
	Message string
}