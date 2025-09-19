package certs

import (
	"crypto/x509"
	"time"
)

// CertDiagnostics contains diagnostic information about certificate validation
type CertDiagnostics struct {
	// Certificate identification
	Subject         string `json:"subject"`
	Issuer          string `json:"issuer"`
	SerialNumber    string `json:"serial_number"`
	
	// Validity information
	NotBefore       time.Time `json:"not_before"`
	NotAfter        time.Time `json:"not_after"`
	IsExpired       bool      `json:"is_expired"`
	IsNotYetValid   bool      `json:"is_not_yet_valid"`
	
	// Chain information
	ChainLength     int    `json:"chain_length"`
	ChainComplete   bool   `json:"chain_complete"`
	TrustedRoot     bool   `json:"trusted_root"`
	
	// Technical details
	SignatureAlgorithm string            `json:"signature_algorithm"`
	PublicKeyAlgorithm string            `json:"public_key_algorithm"`
	KeySize           int               `json:"key_size"`
	KeyUsages         []x509.KeyUsage   `json:"key_usages"`
	ExtKeyUsages      []x509.ExtKeyUsage `json:"ext_key_usages"`
	
	// Validation results
	ValidationErrors []ValidationErrorType `json:"validation_errors"`
	WarningCount     int                   `json:"warning_count"`
	ErrorCount       int                   `json:"error_count"`
}