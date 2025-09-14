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
	"testing"
	"time"
)

func TestCertificateTypeConstants(t *testing.T) {
	types := map[string]string{
		"CA":           CertTypeCA,
		"Client":       CertTypeClient,
		"Server":       CertTypeServer,
		"Intermediate": CertTypeIntermediate,
	}

	expected := map[string]string{
		"CA":           "ca",
		"Client":       "client",
		"Server":       "server",
		"Intermediate": "intermediate",
	}

	for name, actual := range types {
		if actual != expected[name] {
			t.Errorf("Expected %s to be %q, got %q", name, expected[name], actual)
		}
	}
}

func TestCertificateExtensionConstants(t *testing.T) {
	extensions := map[string]string{
		"PEM": CertExtensionPEM,
		"CRT": CertExtensionCRT,
		"CER": CertExtensionCER,
		"KEY": CertExtensionKEY,
		"P12": CertExtensionP12,
		"PFX": CertExtensionPFX,
	}

	expected := map[string]string{
		"PEM": ".pem",
		"CRT": ".crt",
		"CER": ".cer",
		"KEY": ".key",
		"P12": ".p12",
		"PFX": ".pfx",
	}

	for name, actual := range extensions {
		if actual != expected[name] {
			t.Errorf("Expected %s extension to be %q, got %q", name, expected[name], actual)
		}
	}
}

func TestDefaultPathConstants(t *testing.T) {
	paths := []struct {
		name     string
		actual   string
		expected string
	}{
		{"DefaultCertDir", DefaultCertDir, "/etc/ssl/certs"},
		{"DefaultPrivateKeyDir", DefaultPrivateKeyDir, "/etc/ssl/private"},
		{"DefaultCAFile", DefaultCAFile, "/etc/ssl/certs/ca-certificates.crt"},
		{"DefaultClientCertFile", DefaultClientCertFile, "client.crt"},
		{"DefaultClientKeyFile", DefaultClientKeyFile, "client.key"},
		{"DefaultServerCertFile", DefaultServerCertFile, "server.crt"},
		{"DefaultServerKeyFile", DefaultServerKeyFile, "server.key"},
	}

	for _, test := range paths {
		if test.actual != test.expected {
			t.Errorf("Expected %s to be %q, got %q", test.name, test.expected, test.actual)
		}
	}
}

func TestTLSVersionConstants(t *testing.T) {
	if MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinTLSVersion to be TLS 1.2, got %d", MinTLSVersion)
	}

	if PreferredTLSVersion != tls.VersionTLS13 {
		t.Errorf("Expected PreferredTLSVersion to be TLS 1.3, got %d", PreferredTLSVersion)
	}

	if TLSVersion12 != tls.VersionTLS12 {
		t.Errorf("Expected TLSVersion12 to be TLS 1.2, got %d", TLSVersion12)
	}

	if TLSVersion13 != tls.VersionTLS13 {
		t.Errorf("Expected TLSVersion13 to be TLS 1.3, got %d", TLSVersion13)
	}
}

func TestTimeoutConstants(t *testing.T) {
	if DefaultValidationTimeout != 30*time.Second {
		t.Errorf("Expected DefaultValidationTimeout to be 30s, got %v", DefaultValidationTimeout)
	}

	if CertificateExpiryWarningThreshold != 30*24*time.Hour {
		t.Errorf("Expected CertificateExpiryWarningThreshold to be 30 days, got %v", CertificateExpiryWarningThreshold)
	}

	if CertificateMaxAge != 10*365*24*time.Hour {
		t.Errorf("Expected CertificateMaxAge to be 10 years, got %v", CertificateMaxAge)
	}

	if CertificateMinValidityPeriod != 24*time.Hour {
		t.Errorf("Expected CertificateMinValidityPeriod to be 1 day, got %v", CertificateMinValidityPeriod)
	}
}

func TestPEMBlockConstants(t *testing.T) {
	blocks := map[string]string{
		"Certificate":        PEMBlockCertificate,
		"PrivateKey":         PEMBlockPrivateKey,
		"RSAPrivateKey":      PEMBlockRSAPrivateKey,
		"ECPrivateKey":       PEMBlockECPrivateKey,
		"CertificateRequest": PEMBlockCertificateRequest,
		"PublicKey":          PEMBlockPublicKey,
	}

	expected := map[string]string{
		"Certificate":        "CERTIFICATE",
		"PrivateKey":         "PRIVATE KEY",
		"RSAPrivateKey":      "RSA PRIVATE KEY",
		"ECPrivateKey":       "EC PRIVATE KEY",
		"CertificateRequest": "CERTIFICATE REQUEST",
		"PublicKey":          "PUBLIC KEY",
	}

	for name, actual := range blocks {
		if actual != expected[name] {
			t.Errorf("Expected %s PEM block to be %q, got %q", name, expected[name], actual)
		}
	}
}

func TestCertificateFieldConstants(t *testing.T) {
	fields := map[string]string{
		"CommonName":         CommonNameField,
		"Organization":       OrganizationField,
		"OrganizationalUnit": OrganizationalUnitField,
		"Country":            CountryField,
		"State":              StateField,
		"Locality":           LocalityField,
		"EmailAddress":       EmailAddressField,
	}

	expected := map[string]string{
		"CommonName":         "CN",
		"Organization":       "O",
		"OrganizationalUnit": "OU",
		"Country":            "C",
		"State":              "ST",
		"Locality":           "L",
		"EmailAddress":       "emailAddress",
	}

	for name, actual := range fields {
		if actual != expected[name] {
			t.Errorf("Expected %s field to be %q, got %q", name, expected[name], actual)
		}
	}
}

func TestErrorMessageConstants(t *testing.T) {
	errorMessages := []string{
		ErrCertificateExpired,
		ErrCertificateNotYetValid,
		ErrInvalidCertificate,
		ErrCertificateChainInvalid,
		ErrPrivateKeyMismatch,
		ErrInvalidPEMBlock,
		ErrCertificateNotFound,
		ErrPrivateKeyNotFound,
		ErrUnsupportedKeyType,
		ErrInvalidCertificateAuthority,
		ErrCertificateRevoked,
	}

	for _, msg := range errorMessages {
		if msg == "" {
			t.Error("Error message constant should not be empty")
		}
	}
}

func TestPreferredCipherSuites(t *testing.T) {
	if len(PreferredCipherSuites) == 0 {
		t.Error("PreferredCipherSuites should not be empty")
	}

	// Check that we include TLS 1.3 cipher suites
	hasTLS13 := false
	for _, suite := range PreferredCipherSuites {
		if suite == tls.TLS_AES_256_GCM_SHA384 || 
		   suite == tls.TLS_AES_128_GCM_SHA256 || 
		   suite == tls.TLS_CHACHA20_POLY1305_SHA256 {
			hasTLS13 = true
			break
		}
	}

	if !hasTLS13 {
		t.Error("PreferredCipherSuites should include TLS 1.3 cipher suites")
	}
}

func TestSecureTLSCurves(t *testing.T) {
	if len(SecureTLSCurves) == 0 {
		t.Error("SecureTLSCurves should not be empty")
	}

	// Check that we include modern curves
	hasModernCurves := false
	for _, curve := range SecureTLSCurves {
		if curve == tls.X25519 || curve == tls.CurveP256 {
			hasModernCurves = true
			break
		}
	}

	if !hasModernCurves {
		t.Error("SecureTLSCurves should include modern curves like X25519 or P256")
	}
}