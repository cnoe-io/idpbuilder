package fallback

// TrustStoreManager defines the interface for managing certificate trust
// This is a temporary interface until we integrate with the actual certs package
type TrustStoreManager interface {
	// SetInsecure enables or disables insecure mode for a registry
	SetInsecure(registry string, insecure bool) error
}

// Optional extended interfaces that implementations may support
type SystemCertManager interface {
	TrustStoreManager
	// SetUseSystemCerts enables or disables system certificate usage for a registry
	SetUseSystemCerts(registry string, use bool) error
}

type CertificateManager interface {
	TrustStoreManager
	// AddCertificate adds a certificate for a specific registry
	AddCertificate(registry string, certData []byte) error
}