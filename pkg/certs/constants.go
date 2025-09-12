// Copyright 2024 The IDP Builder Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certs

import (
	"crypto/tls"
	"time"
)

// Certificate type identifiers
const (
	// CertTypeCA identifies a Certificate Authority certificate
	CertTypeCA = "ca"
	
	// CertTypeClient identifies a client certificate
	CertTypeClient = "client"
	
	// CertTypeServer identifies a server certificate
	CertTypeServer = "server"
	
	// CertTypeIntermediate identifies an intermediate CA certificate
	CertTypeIntermediate = "intermediate"
)

// Certificate file extensions
const (
	// CertExtensionPEM for PEM-encoded certificates
	CertExtensionPEM = ".pem"
	
	// CertExtensionCRT for certificate files
	CertExtensionCRT = ".crt"
	
	// CertExtensionCER for certificate files (alternative)
	CertExtensionCER = ".cer"
	
	// CertExtensionKEY for private key files
	CertExtensionKEY = ".key"
	
	// CertExtensionP12 for PKCS#12 files
	CertExtensionP12 = ".p12"
	
	// CertExtensionPFX for PKCS#12 files (Windows)
	CertExtensionPFX = ".pfx"
)

// Default certificate paths and locations
const (
	// DefaultCertDir is the default certificate directory
	DefaultCertDir = "/etc/ssl/certs"
	
	// DefaultPrivateKeyDir is the default private key directory
	DefaultPrivateKeyDir = "/etc/ssl/private"
	
	// DefaultCAFile is the default CA bundle file location
	DefaultCAFile = "/etc/ssl/certs/ca-certificates.crt"
	
	// DefaultClientCertFile is the default client certificate file
	DefaultClientCertFile = "client.crt"
	
	// DefaultClientKeyFile is the default client private key file
	DefaultClientKeyFile = "client.key"
	
	// DefaultServerCertFile is the default server certificate file
	DefaultServerCertFile = "server.crt"
	
	// DefaultServerKeyFile is the default server private key file
	DefaultServerKeyFile = "server.key"
)

// TLS version constants for minimum security requirements
const (
	// MinTLSVersion defines the minimum acceptable TLS version
	MinTLSVersion = tls.VersionTLS12
	
	// PreferredTLSVersion defines the preferred TLS version
	PreferredTLSVersion = tls.VersionTLS13
	
	// TLSVersion12 constant for TLS 1.2
	TLSVersion12 = tls.VersionTLS12
	
	// TLSVersion13 constant for TLS 1.3
	TLSVersion13 = tls.VersionTLS13
)

// Certificate validation timeouts and thresholds
const (
	// DefaultValidationTimeout for certificate validation operations
	DefaultValidationTimeout = 30 * time.Second
	
	// CertificateExpiryWarningThreshold when to warn about expiring certificates
	CertificateExpiryWarningThreshold = 30 * 24 * time.Hour // 30 days
	
	// CertificateMaxAge maximum age for accepting certificates
	CertificateMaxAge = 10 * 365 * 24 * time.Hour // 10 years
	
	// CertificateMinValidityPeriod minimum validity period for new certificates
	CertificateMinValidityPeriod = 24 * time.Hour // 1 day
)

// PEM block type constants
const (
	// PEMBlockCertificate for certificate PEM blocks
	PEMBlockCertificate = "CERTIFICATE"
	
	// PEMBlockPrivateKey for private key PEM blocks
	PEMBlockPrivateKey = "PRIVATE KEY"
	
	// PEMBlockRSAPrivateKey for RSA private key PEM blocks
	PEMBlockRSAPrivateKey = "RSA PRIVATE KEY"
	
	// PEMBlockECPrivateKey for EC private key PEM blocks
	PEMBlockECPrivateKey = "EC PRIVATE KEY"
	
	// PEMBlockCertificateRequest for certificate request PEM blocks
	PEMBlockCertificateRequest = "CERTIFICATE REQUEST"
	
	// PEMBlockPublicKey for public key PEM blocks
	PEMBlockPublicKey = "PUBLIC KEY"
)

// Common certificate field names and identifiers
const (
	// CommonNameField identifies the Common Name field
	CommonNameField = "CN"
	
	// OrganizationField identifies the Organization field
	OrganizationField = "O"
	
	// OrganizationalUnitField identifies the Organizational Unit field
	OrganizationalUnitField = "OU"
	
	// CountryField identifies the Country field
	CountryField = "C"
	
	// StateField identifies the State/Province field
	StateField = "ST"
	
	// LocalityField identifies the Locality/City field
	LocalityField = "L"
	
	// EmailAddressField identifies the Email Address field
	EmailAddressField = "emailAddress"
)

// Certificate validation error messages
const (
	// ErrCertificateExpired when certificate has expired
	ErrCertificateExpired = "certificate has expired"
	
	// ErrCertificateNotYetValid when certificate is not yet valid
	ErrCertificateNotYetValid = "certificate is not yet valid"
	
	// ErrInvalidCertificate when certificate format is invalid
	ErrInvalidCertificate = "invalid certificate format"
	
	// ErrCertificateChainInvalid when certificate chain validation fails
	ErrCertificateChainInvalid = "certificate chain validation failed"
	
	// ErrPrivateKeyMismatch when private key doesn't match certificate
	ErrPrivateKeyMismatch = "private key does not match certificate"
	
	// ErrInvalidPEMBlock when PEM block format is invalid
	ErrInvalidPEMBlock = "invalid PEM block format"
	
	// ErrCertificateNotFound when certificate file is not found
	ErrCertificateNotFound = "certificate not found"
	
	// ErrPrivateKeyNotFound when private key file is not found
	ErrPrivateKeyNotFound = "private key not found"
	
	// ErrUnsupportedKeyType when key type is not supported
	ErrUnsupportedKeyType = "unsupported private key type"
	
	// ErrInvalidCertificateAuthority when CA certificate is invalid
	ErrInvalidCertificateAuthority = "invalid certificate authority"
	
	// ErrCertificateRevoked when certificate has been revoked
	ErrCertificateRevoked = "certificate has been revoked"
)

// TLS cipher suite preferences for secure configurations
var (
	// PreferredCipherSuites defines the preferred cipher suites for TLS
	PreferredCipherSuites = []uint16{
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
	
	// SecureTLSCurves defines the preferred elliptic curves for TLS
	SecureTLSCurves = []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
		tls.CurveP384,
	}
)