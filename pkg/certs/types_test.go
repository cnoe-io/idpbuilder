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
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"
)

// Helper function to create a test certificate PEM
func createTestCertPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

const invalidCertPEM = `-----BEGIN CERTIFICATE-----
InvalidCertificateData
-----END CERTIFICATE-----`

// NOTE: Private helper functions removed - now use exported CreateTestCertificate() and CreateExpiredTestCertificate() from test_helpers.go

func TestNewCertificate(t *testing.T) {
	x509Cert := CreateTestCertificate(t)
	pemData := createTestCertPEM(x509Cert)
	
	cert := NewCertificate(x509Cert, pemData)
	
	if cert.Certificate != x509Cert {
		t.Error("Certificate should contain the provided x509.Certificate")
	}
	
	if string(cert.PEMData) != string(pemData) {
		t.Error("PEMData should match the provided PEM data")
	}
}

func TestCertificateIsValid(t *testing.T) {
	// Test valid certificate
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	if !validCert.IsValid() {
		t.Error("Valid certificate should return true for IsValid()")
	}

	// Test expired certificate
	expiredX509Cert := CreateExpiredTestCertificate(t)
	expiredCert := NewCertificate(expiredX509Cert, createTestCertPEM(expiredX509Cert))
	if expiredCert.IsValid() {
		t.Error("Expired certificate should return false for IsValid()")
	}
}

func TestCertificateIsExpired(t *testing.T) {
	// Test valid certificate
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	if validCert.IsExpired() {
		t.Error("Valid certificate should return false for IsExpired()")
	}

	// Test expired certificate
	expiredX509Cert := CreateExpiredTestCertificate(t)
	expiredCert := NewCertificate(expiredX509Cert, createTestCertPEM(expiredX509Cert))
	if !expiredCert.IsExpired() {
		t.Error("Expired certificate should return true for IsExpired()")
	}
}

func TestCertificateWillExpireSoon(t *testing.T) {
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	
	// Should not expire within 1 hour
	if validCert.WillExpireSoon(time.Hour) {
		t.Error("Certificate should not expire within 1 hour")
	}
	
	// Should expire within 2 years
	if !validCert.WillExpireSoon(2 * 365 * 24 * time.Hour) {
		t.Error("Certificate should expire within 2 years")
	}
}

func TestCertificateGetCommonName(t *testing.T) {
	x509Cert := CreateTestCertificate(t)
	cert := NewCertificate(x509Cert, createTestCertPEM(x509Cert))
	
	if cert.GetCommonName() != "test.example.com" {
		t.Errorf("Expected common name 'test.example.com', got '%s'", cert.GetCommonName())
	}
}

func TestCertificateGetSANs(t *testing.T) {
	x509Cert := CreateTestCertificate(t)
	cert := NewCertificate(x509Cert, createTestCertPEM(x509Cert))
	
	sans := cert.GetSANs()
	if len(sans) != 1 || sans[0] != "test.example.com" {
		t.Errorf("Expected SANs ['test.example.com'], got %v", sans)
	}
}

func TestNewCertificateBundle(t *testing.T) {
	bundle := NewCertificateBundle()
	
	if bundle == nil {
		t.Fatal("NewCertificateBundle should return a non-nil bundle")
	}
	
	if bundle.CertificateChain == nil {
		t.Error("CertificateChain should be initialized")
	}
	
	if len(bundle.CertificateChain) != 0 {
		t.Error("CertificateChain should be empty initially")
	}
}

func TestCertificateBundleHasPrivateKey(t *testing.T) {
	bundle := NewCertificateBundle()
	
	if bundle.HasPrivateKey() {
		t.Error("Empty bundle should not have private key")
	}
	
	bundle.ClientKey = "dummy-key"
	if !bundle.HasPrivateKey() {
		t.Error("Bundle with client key should have private key")
	}
}

func TestCertificateBundleIsValid(t *testing.T) {
	bundle := NewCertificateBundle()
	
	// Empty bundle should be valid
	if !bundle.IsValid() {
		t.Error("Empty bundle should be valid")
	}
	
	// Bundle with valid CA cert
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	bundle.CACert = validCert
	if !bundle.IsValid() {
		t.Error("Bundle with valid CA cert should be valid")
	}
	
	// Bundle with expired CA cert
	expiredX509Cert := CreateExpiredTestCertificate(t)
	expiredCert := NewCertificate(expiredX509Cert, createTestCertPEM(expiredX509Cert))
	bundle.CACert = expiredCert
	if bundle.IsValid() {
		t.Error("Bundle with expired CA cert should be invalid")
	}
}

func TestCertificateBundleUpdateValidityPeriod(t *testing.T) {
	bundle := NewCertificateBundle()
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	
	bundle.CACert = validCert
	bundle.UpdateValidityPeriod()
	
	if bundle.ValidFrom.IsZero() {
		t.Error("ValidFrom should be set after UpdateValidityPeriod")
	}
	
	if bundle.ValidUntil.IsZero() {
		t.Error("ValidUntil should be set after UpdateValidityPeriod")
	}
}

func TestNewTLSConfig(t *testing.T) {
	config := NewTLSConfig()
	
	if config == nil {
		t.Fatal("NewTLSConfig should return a non-nil config")
	}
	
	if config.MinVersion != MinTLSVersion {
		t.Errorf("Expected MinVersion to be %d, got %d", MinTLSVersion, config.MinVersion)
	}
	
	if config.MaxVersion != PreferredTLSVersion {
		t.Errorf("Expected MaxVersion to be %d, got %d", PreferredTLSVersion, config.MaxVersion)
	}
	
	if len(config.CipherSuites) != len(PreferredCipherSuites) {
		t.Error("CipherSuites should match PreferredCipherSuites")
	}
	
	if len(config.CurvePreferences) != len(SecureTLSCurves) {
		t.Error("CurvePreferences should match SecureTLSCurves")
	}
	
	if config.ClientAuth != tls.NoClientCert {
		t.Error("Default ClientAuth should be NoClientCert")
	}
}

func TestTLSConfigToTLSConfig(t *testing.T) {
	config := NewTLSConfig()
	config.ServerName = "test.example.com"
	config.InsecureSkipVerify = true
	
	tlsConfig := config.ToTLSConfig()
	
	if tlsConfig.ServerName != config.ServerName {
		t.Error("ServerName should be preserved")
	}
	
	if tlsConfig.InsecureSkipVerify != config.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be preserved")
	}
	
	if tlsConfig.MinVersion != config.MinVersion {
		t.Error("MinVersion should be preserved")
	}
}

func TestTLSConfigAddClientCertificate(t *testing.T) {
	config := NewTLSConfig()
	
	initialCount := len(config.Certificates)
	
	cert := tls.Certificate{}
	config.AddClientCertificate(cert)
	
	if len(config.Certificates) != initialCount+1 {
		t.Error("Certificate should be added to Certificates slice")
	}
}

func TestTLSConfigSetRootCA(t *testing.T) {
	config := NewTLSConfig()
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	
	err := config.SetRootCA(validCert)
	if err != nil {
		t.Fatalf("SetRootCA should not return error: %v", err)
	}
	
	if config.RootCAs == nil {
		t.Error("RootCAs should be initialized")
	}
	
	// Test with nil certificate
	err = config.SetRootCA(&Certificate{})
	if err == nil {
		t.Error("SetRootCA should return error for invalid certificate")
	}
}

func TestNewBasicCertificateValidator(t *testing.T) {
	caCerts := x509.NewCertPool()
	validator := NewBasicCertificateValidator(caCerts)
	
	if validator == nil {
		t.Fatal("NewBasicCertificateValidator should return non-nil validator")
	}
	
	if validator.TrustedCAs != caCerts {
		t.Error("TrustedCAs should be set to provided CA pool")
	}
	
	if !validator.CheckExpiry {
		t.Error("CheckExpiry should be enabled by default")
	}
}

func TestBasicCertificateValidatorValidate(t *testing.T) {
	validator := NewBasicCertificateValidator(nil)
	
	// Test nil certificate
	err := validator.Validate(nil)
	if err == nil {
		t.Error("Validate should return error for nil certificate")
	}
	
	// Test invalid certificate
	err = validator.Validate(&Certificate{})
	if err == nil {
		t.Error("Validate should return error for invalid certificate")
	}
	
	// Test valid certificate
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	err = validator.Validate(validCert)
	if err != nil {
		t.Errorf("Validate should not return error for valid certificate: %v", err)
	}
	
	// Test expired certificate
	expiredX509Cert := CreateExpiredTestCertificate(t)
	expiredCert := NewCertificate(expiredX509Cert, createTestCertPEM(expiredX509Cert))
	err = validator.Validate(expiredCert)
	if err == nil {
		t.Error("Validate should return error for expired certificate")
	}
}

func TestBasicCertificateValidatorValidateChain(t *testing.T) {
	validator := NewBasicCertificateValidator(nil)
	
	// Test empty chain
	err := validator.ValidateChain([]*Certificate{})
	if err == nil {
		t.Error("ValidateChain should return error for empty chain")
	}
	
	// Test chain with valid certificate
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	err = validator.ValidateChain([]*Certificate{validCert})
	if err != nil {
		t.Errorf("ValidateChain should not return error for valid chain: %v", err)
	}
	
	// Test chain with expired certificate
	expiredX509Cert := CreateExpiredTestCertificate(t)
	expiredCert := NewCertificate(expiredX509Cert, createTestCertPEM(expiredX509Cert))
	err = validator.ValidateChain([]*Certificate{expiredCert})
	if err == nil {
		t.Error("ValidateChain should return error for chain with expired certificate")
	}
}

func TestBasicCertificateValidatorIsExpired(t *testing.T) {
	validator := NewBasicCertificateValidator(nil)
	
	// Test nil certificate
	if !validator.IsExpired(nil) {
		t.Error("IsExpired should return true for nil certificate")
	}
	
	// Test valid certificate
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	if validator.IsExpired(validCert) {
		t.Error("IsExpired should return false for valid certificate")
	}
	
	// Test expired certificate
	expiredX509Cert := CreateExpiredTestCertificate(t)
	expiredCert := NewCertificate(expiredX509Cert, createTestCertPEM(expiredX509Cert))
	if !validator.IsExpired(expiredCert) {
		t.Error("IsExpired should return true for expired certificate")
	}
}

func TestBasicCertificateValidatorWillExpireSoon(t *testing.T) {
	validator := NewBasicCertificateValidator(nil)
	
	// Test nil certificate
	if !validator.WillExpireSoon(nil, time.Hour) {
		t.Error("WillExpireSoon should return true for nil certificate")
	}
	
	// Test valid certificate
	validX509Cert := CreateTestCertificate(t)
	validCert := NewCertificate(validX509Cert, createTestCertPEM(validX509Cert))
	if validator.WillExpireSoon(validCert, time.Hour) {
		t.Error("WillExpireSoon should return false for certificate not expiring soon")
	}
	
	if !validator.WillExpireSoon(validCert, 2*365*24*time.Hour) {
		t.Error("WillExpireSoon should return true for certificate expiring within 2 years")
	}
}

func TestParseCertificateFromPEM(t *testing.T) {
	// Test invalid PEM data
	_, err := ParseCertificateFromPEM([]byte("invalid pem data"))
	if err == nil {
		t.Error("ParseCertificateFromPEM should return error for invalid PEM")
	}
	
	// Test invalid certificate data
	_, err = ParseCertificateFromPEM([]byte(invalidCertPEM))
	if err == nil {
		t.Error("ParseCertificateFromPEM should return error for invalid certificate")
	}
	
	// Test wrong PEM block type
	keyPEM := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC659HGH
-----END PRIVATE KEY-----`
	_, err = ParseCertificateFromPEM([]byte(keyPEM))
	if err == nil {
		t.Error("ParseCertificateFromPEM should return error for private key PEM")
	}
}

func TestParseCertificatesFromPEM(t *testing.T) {
	// Test with no certificates
	_, err := ParseCertificatesFromPEM([]byte("no certificates here"))
	if err == nil {
		t.Error("ParseCertificatesFromPEM should return error when no certificates found")
	}
	
	// Test with multiple certificates (using same cert twice for simplicity)
	testX509Cert := CreateTestCertificate(t)
	testPEM := createTestCertPEM(testX509Cert)
	multiCertPEM := string(testPEM) + "\n" + string(testPEM)
	certs, err := ParseCertificatesFromPEM([]byte(multiCertPEM))
	if err != nil {
		t.Errorf("ParseCertificatesFromPEM should not return error: %v", err)
	}
	
	if len(certs) != 2 {
		t.Errorf("Expected 2 certificates, got %d", len(certs))
	}
}

func TestLoadCertificateFromReader(t *testing.T) {
	testX509Cert := CreateTestCertificate(t)
	testPEM := createTestCertPEM(testX509Cert)
	reader := strings.NewReader(string(testPEM))
	
	cert, err := LoadCertificateFromReader(reader)
	if err != nil {
		t.Errorf("LoadCertificateFromReader should not return error: %v", err)
	}
	
	if cert == nil {
		t.Error("LoadCertificateFromReader should return non-nil certificate")
	}
}

func TestCertificateError(t *testing.T) {
	// Test error without cause
	err := &CertificateError{
		Type:    "test-error",
		Message: "test message",
	}
	
	if err.Error() != "test message" {
		t.Errorf("Expected 'test message', got '%s'", err.Error())
	}
	
	if err.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause is set")
	}
	
	// Test error with cause
	cause := &CertificateError{Type: "cause-error", Message: "cause message"}
	err = &CertificateError{
		Type:    "wrapper-error",
		Message: "wrapper message",
		Cause:   cause,
	}
	
	expected := "wrapper message: cause message"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
	
	if err.Unwrap() != cause {
		t.Error("Unwrap should return the cause error")
	}
}