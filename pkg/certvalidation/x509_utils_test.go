package certvalidation

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"strings"
	"testing"
	"time"
)

const testCertPEM = `-----BEGIN CERTIFICATE-----
MIIDFzCCAf+gAwIBAgIUCKxQt+/W7ZIkHUiSLo7BVSUvk6MwDQYJKoZIhvcNAQEL
BQAwGzEZMBcGA1UEAwwQVGVzdCBDZXJ0aWZpY2F0ZTAeFw0yNTA5MDcxOTM5NTVa
Fw0yNjA5MDcxOTM5NTVaMBsxGTAXBgNVBAMMEFRlc3QgQ2VydGlmaWNhdGUwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCSLU8cpbV31v8cD6vc1rwUw/Xg
+4Hyx3ZU8PXjLOzovIzYdPNqZ9dZkFfxDFwVkJ5Y9TEudkoboPiR/WxpN7wJM11G
ZuSBd1tRnGFNgnJINxNigMNSD+VVi0gGJj75600spoHKFtpWTYt0+OZF6zijIFCH
QhEMsUDNwbgbXf6DHt7s18SvNjS4pACWfG4N08IWNlMoQnr9O6R123LTwz3z7OC7
/pf2LCM1yMTwkTz9EZNxARcIJRErv6HoCNfLT7NZ2qJ8VIcIxThRS+cuQu+TEyOH
F7N8lJuV7VDixxqDfx+BRaohcLeTyXyiqp1RExYJAs3tbkUgy0b8BZLIM9C5AgMB
AAGjUzBRMB0GA1UdDgQWBBT0MH41W+RZOdDNIP/771Iauhw02zAfBgNVHSMEGDAW
gBT0MH41W+RZOdDNIP/771Iauhw02zAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3
DQEBCwUAA4IBAQAcvzJSsBmFjG2woBJCPYd69Pl0iKXhkH9LOY3XSyZdxNISUd10
ytLttS8NdsWKdd0gfFw6P7y/FLds0HfWrVr8sz3yJAAbubx95r11g89PH7O12Fc3
g/wmfjKx87nSSwAb6U6Lwh88RYZYShrwSGGhIiV8lJ99MSUTB1OeJGPReRDoFtLr
0hlxlY0Rc9S/nyhBVi+uLsR+cWufBZQN1a1jsaeK9uSmSzEnpgcBUVOoHV20oVba
zp6C10QOR4vHtcbRIkTtxOvjxGsYE1kXn/KbdtuBt/uVJxuyzwe8H6apHekwHSgI
RzRXRAqIazcIK73G6N1tHFZ+sYK6DPJhcMb8
-----END CERTIFICATE-----`

const testMultipleCertsPEM = `-----BEGIN CERTIFICATE-----
MIIDFzCCAf+gAwIBAgIUCKxQt+/W7ZIkHUiSLo7BVSUvk6MwDQYJKoZIhvcNAQEL
BQAwGzEZMBcGA1UEAwwQVGVzdCBDZXJ0aWZpY2F0ZTAeFw0yNTA5MDcxOTM5NTVa
Fw0yNjA5MDcxOTM5NTVaMBsxGTAXBgNVBAMMEFRlc3QgQ2VydGlmaWNhdGUwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCSLU8cpbV31v8cD6vc1rwUw/Xg
+4Hyx3ZU8PXjLOzovIzYdPNqZ9dZkFfxDFwVkJ5Y9TEudkoboPiR/WxpN7wJM11G
ZuSBd1tRnGFNgnJINxNigMNSD+VVi0gGJj75600spoHKFtpWTYt0+OZF6zijIFCH
QhEMsUDNwbgbXf6DHt7s18SvNjS4pACWfG4N08IWNlMoQnr9O6R123LTwz3z7OC7
/pf2LCM1yMTwkTz9EZNxARcIJRErv6HoCNfLT7NZ2qJ8VIcIxThRS+cuQu+TEyOH
F7N8lJuV7VDixxqDfx+BRaohcLeTyXyiqp1RExYJAs3tbkUgy0b8BZLIM9C5AgMB
AAGjUzBRMB0GA1UdDgQWBBT0MH41W+RZOdDNIP/771Iauhw02zAfBgNVHSMEGDAW
gBT0MH41W+RZOdDNIP/771Iauhw02zAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3
DQEBCwUAA4IBAQAcvzJSsBmFjG2woBJCPYd69Pl0iKXhkH9LOY3XSyZdxNISUd10
ytLttS8NdsWKdd0gfFw6P7y/FLds0HfWrVr8sz3yJAAbubx95r11g89PH7O12Fc3
g/wmfjKx87nSSwAb6U6Lwh88RYZYShrwSGGhIiV8lJ99MSUTB1OeJGPReRDoFtLr
0hlxlY0Rc9S/nyhBVi+uLsR+cWufBZQN1a1jsaeK9uSmSzEnpgcBUVOoHV20oVba
zp6C10QOR4vHtcbRIkTtxOvjxGsYE1kXn/KbdtuBt/uVJxuyzwe8H6apHekwHSgI
RzRXRAqIazcIK73G6N1tHFZ+sYK6DPJhcMb8
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDFzCCAf+gAwIBAgIUCKxQt+/W7ZIkHUiSLo7BVSUvk6MwDQYJKoZIhvcNAQEL
BQAwGzEZMBcGA1UEAwwQVGVzdCBDZXJ0aWZpY2F0ZTAeFw0yNTA5MDcxOTM5NTVa
Fw0yNjA5MDcxOTM5NTVaMBsxGTAXBgNVBAMMEFRlc3QgQ2VydGlmaWNhdGUwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCSLU8cpbV31v8cD6vc1rwUw/Xg
+4Hyx3ZU8PXjLOzovIzYdPNqZ9dZkFfxDFwVkJ5Y9TEudkoboPiR/WxpN7wJM11G
ZuSBd1tRnGFNgnJINxNigMNSD+VVi0gGJj75600spoHKFtpWTYt0+OZF6zijIFCH
QhEMsUDNwbgbXf6DHt7s18SvNjS4pACWfG4N08IWNlMoQnr9O6R123LTwz3z7OC7
/pf2LCM1yMTwkTz9EZNxARcIJRErv6HoCNfLT7NZ2qJ8VIcIxThRS+cuQu+TEyOH
F7N8lJuV7VDixxqDfx+BRaohcLeTyXyiqp1RExYJAs3tbkUgy0b8BZLIM9C5AgMB
AAGjUzBRMB0GA1UdDgQWBBT0MH41W+RZOdDNIP/771Iauhw02zAfBgNVHSMEGDAW
gBT0MH41W+RZOdDNIP/771Iauhw02zAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3
DQEBCwUAA4IBAQAcvzJSsBmFjG2woBJCPYd69Pl0iKXhkH9LOY3XSyZdxNISUd10
ytLttS8NdsWKdd0gfFw6P7y/FLds0HfWrVr8sz3yJAAbubx95r11g89PH7O12Fc3
g/wmfjKx87nSSwAb6U6Lwh88RYZYShrwSGGhIiV8lJ99MSUTB1OeJGPReRDoFtLr
0hlxlY0Rc9S/nyhBVi+uLsR+cWufBZQN1a1jsaeK9uSmSzEnpgcBUVOoHV20oVba
zp6C10QOR4vHtcbRIkTtxOvjxGsYE1kXn/KbdtuBt/uVJxuyzwe8H6apHekwHSgI
RzRXRAqIazcIK73G6N1tHFZ+sYK6DPJhcMb8
-----END CERTIFICATE-----`

func TestParsePEMCertificate(t *testing.T) {
	// Test valid PEM
	cert, err := ParsePEMCertificate([]byte(testCertPEM))
	if err != nil {
		t.Errorf("Expected successful parsing, got error: %v", err)
	}
	if cert == nil {
		t.Error("Expected non-nil certificate")
	}

	// Test invalid PEM
	invalidPEM := []byte("not a pem certificate")
	cert, err = ParsePEMCertificate(invalidPEM)
	if err == nil {
		t.Error("Expected error for invalid PEM")
	}
	if cert != nil {
		t.Error("Expected nil certificate for invalid PEM")
	}

	// Test wrong block type
	wrongTypePEM := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKB
wko6tRCnXlGIKzgx6tNlpUYEhSYRFM+qUj4p+C/xFoVoUpVtgf2j5YnxRm4N+rNa
-----END PRIVATE KEY-----`
	cert, err = ParsePEMCertificate([]byte(wrongTypePEM))
	if err == nil {
		t.Error("Expected error for wrong block type")
	}
}

func TestParsePEMCertificates(t *testing.T) {
	// Test multiple certificates
	certs, err := ParsePEMCertificates([]byte(testMultipleCertsPEM))
	if err != nil {
		t.Errorf("Expected successful parsing, got error: %v", err)
	}
	if len(certs) != 2 {
		t.Errorf("Expected 2 certificates, got %d", len(certs))
	}

	// Test empty PEM
	certs, err = ParsePEMCertificates([]byte(""))
	if err == nil {
		t.Error("Expected error for empty PEM")
	}

	// Test no certificates found
	noCertsPEM := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKB
-----END PRIVATE KEY-----`
	certs, err = ParsePEMCertificates([]byte(noCertsPEM))
	if err == nil {
		t.Error("Expected error when no certificates found")
	}
}

func TestCertificateToPEM(t *testing.T) {
	// Create test certificate
	cert, err := ParsePEMCertificate([]byte(testCertPEM))
	if err != nil {
		t.Fatalf("Failed to parse test certificate: %v", err)
	}

	// Convert to PEM
	pemBytes, err := CertificateToPEM(cert)
	if err != nil {
		t.Errorf("Expected successful PEM conversion, got error: %v", err)
	}

	if len(pemBytes) == 0 {
		t.Error("Expected non-empty PEM bytes")
	}

	// Verify it contains PEM headers
	pemString := string(pemBytes)
	if !strings.Contains(pemString, "BEGIN CERTIFICATE") {
		t.Error("Expected PEM to contain BEGIN CERTIFICATE header")
	}

	// Test with nil certificate
	pemBytes, err = CertificateToPEM(nil)
	if err == nil {
		t.Error("Expected error for nil certificate")
	}
}

func TestGetCertificateInfo(t *testing.T) {
	// Test with nil certificate
	info := GetCertificateInfo(nil)
	if info.Subject != "" {
		t.Error("Expected empty info for nil certificate")
	}

	// Create test certificate with specific properties
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"test.example.com", "www.test.example.com"},
		IsCA:         false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	info = GetCertificateInfo(cert)

	if info.CommonName != "test.example.com" {
		t.Errorf("Expected CommonName 'test.example.com', got '%s'", info.CommonName)
	}

	if info.SerialNumber != "12345" {
		t.Errorf("Expected SerialNumber '12345', got '%s'", info.SerialNumber)
	}

	if !info.IsValid {
		t.Error("Expected certificate to be valid")
	}

	if len(info.DNSNames) != 2 {
		t.Errorf("Expected 2 DNS names, got %d", len(info.DNSNames))
	}

	if info.IsCA {
		t.Error("Expected certificate to not be CA")
	}
}

func TestValidateCertificateTime(t *testing.T) {
	// Test with nil certificate
	err := ValidateCertificateTime(nil, time.Now())
	if err == nil {
		t.Error("Expected error for nil certificate")
	}

	// Create certificate with specific validity period
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    now.Add(-time.Hour),     // Valid from 1 hour ago
		NotAfter:     now.Add(time.Hour * 24), // Valid until 24 hours from now
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Test valid time
	err = ValidateCertificateTime(cert, now)
	if err != nil {
		t.Errorf("Expected valid time, got error: %v", err)
	}

	// Test time before validity
	err = ValidateCertificateTime(cert, now.Add(-time.Hour*2))
	if err == nil {
		t.Error("Expected error for time before validity")
	}

	// Test time after validity
	err = ValidateCertificateTime(cert, now.Add(time.Hour*25))
	if err == nil {
		t.Error("Expected error for time after validity")
	}
}

func TestIsSelfSigned(t *testing.T) {
	// Test with nil certificate
	if IsSelfSigned(nil) {
		t.Error("Expected false for nil certificate")
	}

	// Create self-signed certificate
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Self Signed"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
	}

	// Self-signed: template signs itself
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	if !IsSelfSigned(cert) {
		t.Error("Expected certificate to be self-signed")
	}
}

func TestGetCertificateFingerprint(t *testing.T) {
	// Test with nil certificate
	fingerprint := GetCertificateFingerprint(nil)
	if fingerprint != "" {
		t.Error("Expected empty fingerprint for nil certificate")
	}

	// Test with valid certificate
	cert, err := ParsePEMCertificate([]byte(testCertPEM))
	if err != nil {
		t.Fatalf("Failed to parse test certificate: %v", err)
	}

	fingerprint = GetCertificateFingerprint(cert)
	if fingerprint == "" {
		t.Error("Expected non-empty fingerprint")
	}

	if !strings.HasPrefix(fingerprint, "sha256:") {
		t.Error("Expected fingerprint to start with sha256:")
	}
}

func TestFilterCertificatesByUsage(t *testing.T) {
	// Create certificates with different key usages
	key1, _ := rsa.GenerateKey(rand.Reader, 2048)
	key2, _ := rsa.GenerateKey(rand.Reader, 2048)

	template1 := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Digital Signature"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	template2 := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Key Encipherment"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:     x509.KeyUsageKeyEncipherment,
	}

	cert1DER, _ := x509.CreateCertificate(rand.Reader, &template1, &template1, &key1.PublicKey, key1)
	cert2DER, _ := x509.CreateCertificate(rand.Reader, &template2, &template2, &key2.PublicKey, key2)

	cert1, _ := x509.ParseCertificate(cert1DER)
	cert2, _ := x509.ParseCertificate(cert2DER)

	certs := []*x509.Certificate{cert1, cert2}

	// Filter for digital signature
	filtered := FilterCertificatesByUsage(certs, x509.KeyUsageDigitalSignature)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 certificate with digital signature usage, got %d", len(filtered))
	}

	// Filter for key encipherment
	filtered = FilterCertificatesByUsage(certs, x509.KeyUsageKeyEncipherment)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 certificate with key encipherment usage, got %d", len(filtered))
	}

	// Filter for cert sign (none should match)
	filtered = FilterCertificatesByUsage(certs, x509.KeyUsageCertSign)
	if len(filtered) != 0 {
		t.Errorf("Expected 0 certificates with cert sign usage, got %d", len(filtered))
	}
}

func TestSortCertificatesByExpiry(t *testing.T) {
	// Test empty slice
	sorted := SortCertificatesByExpiry([]*x509.Certificate{})
	if len(sorted) != 0 {
		t.Error("Expected empty slice to remain empty")
	}

	// Test single certificate
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24),
	}

	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	cert1, _ := x509.ParseCertificate(certDER)

	sorted = SortCertificatesByExpiry([]*x509.Certificate{cert1})
	if len(sorted) != 1 || sorted[0] != cert1 {
		t.Error("Expected single certificate to remain unchanged")
	}
}

func TestExtractCertificateChainFromPEM(t *testing.T) {
	// Test single certificate
	chain, err := ExtractCertificateChainFromPEM([]byte(testCertPEM))
	if err != nil {
		t.Errorf("Expected successful extraction, got error: %v", err)
	}
	if len(chain) != 1 {
		t.Errorf("Expected 1 certificate, got %d", len(chain))
	}

	// Test multiple certificates
	chain, err = ExtractCertificateChainFromPEM([]byte(testMultipleCertsPEM))
	if err != nil {
		t.Errorf("Expected successful extraction, got error: %v", err)
	}
	if len(chain) != 2 {
		t.Errorf("Expected 2 certificates, got %d", len(chain))
	}

	// Test invalid PEM
	chain, err = ExtractCertificateChainFromPEM([]byte("invalid pem"))
	if err == nil {
		t.Error("Expected error for invalid PEM")
	}
}

func TestFindCertificatesBySubject(t *testing.T) {
	// Create test certificates with different subjects
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	template1 := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24),
	}

	template2 := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "other.example.com"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24),
	}

	cert1DER, _ := x509.CreateCertificate(rand.Reader, &template1, &template1, &key.PublicKey, key)
	cert2DER, _ := x509.CreateCertificate(rand.Reader, &template2, &template2, &key.PublicKey, key)

	cert1, _ := x509.ParseCertificate(cert1DER)
	cert2, _ := x509.ParseCertificate(cert2DER)

	certs := []*x509.Certificate{cert1, cert2}

	// Find by exact common name
	matches := FindCertificatesBySubject(certs, "test.example.com")
	if len(matches) != 1 {
		t.Errorf("Expected 1 match for 'test.example.com', got %d", len(matches))
	}

	// Find by partial subject
	matches = FindCertificatesBySubject(certs, "example.com")
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches for 'example.com', got %d", len(matches))
	}

	// Find non-existent subject
	matches = FindCertificatesBySubject(certs, "nonexistent.com")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for 'nonexistent.com', got %d", len(matches))
	}
}
