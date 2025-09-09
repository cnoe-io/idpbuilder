package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// GetRemoteOptions returns configured remote options for registry operations.
// Integrates with Phase 1 certificate infrastructure for TLS configuration,
// authentication setup, and transport customization.
func (r *giteaRegistryImpl) GetRemoteOptions() []remote.Option {
	if err := r.validateRegistry(); err != nil {
		log.Printf("Warning: Registry validation failed, using basic options: %v", err)
		return r.getBasicOptions()
	}
	
	var options []remote.Option
	
	// Add authentication if available
	if authOption := r.getAuthOption(); authOption != nil {
		options = append(options, authOption)
	}
	
	// Add transport configuration with TLS handling
	if transportOption := r.getTransportOption(); transportOption != nil {
		options = append(options, transportOption)
	}
	
	// Add context timeout
	ctx, cancel := context.WithTimeout(context.Background(), r.getTimeout())
	_ = cancel // Will be used by caller
	options = append(options, remote.WithContext(ctx))
	
	// Add user agent
	options = append(options, remote.WithUserAgent("idpbuilder-oci/gitea-client"))
	
	log.Printf("Configured %d remote options for registry operations", len(options))
	return options
}

// getAuthOption creates authentication option using stored credentials
func (r *giteaRegistryImpl) getAuthOption() remote.Option {
	if r.authn == nil {
		log.Printf("No authenticator available")
		return nil
	}
	
	authenticator := &remoteAuthenticator{
		username: r.authn.username,
		password: r.authn.password,
		token:    r.authn.token,
	}
	
	return remote.WithAuth(authenticator)
}

// getTransportOption creates transport option with Phase 1 certificate integration
func (r *giteaRegistryImpl) getTransportOption() remote.Option {
	transport := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	
	// Configure TLS using Phase 1 infrastructure
	tlsConfig, err := r.configureTLS()
	if err != nil {
		log.Printf("Warning: TLS configuration failed, using default: %v", err)
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: r.config.Insecure,
		}
	} else {
		transport.TLSClientConfig = tlsConfig
	}
	
	return remote.WithTransport(transport)
}

// configureTLS sets up TLS configuration using Phase 1 certificate infrastructure
func (r *giteaRegistryImpl) configureTLS() (*tls.Config, error) {
	// Handle insecure mode using fallback handler
	if r.config.Insecure {
		log.Printf("Using insecure mode for registry connection")
		
		// Use Phase 1 fallback handler to manage insecure connections safely
		r.fallback.SetInsecureMode(true)
		
		return &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         r.baseURL.Host,
		}, nil
	}
	
	// Use Phase 1 trust store for certificate validation
	certPool, err := x509.SystemCertPool()
	if err != nil {
		certPool = x509.NewCertPool()
	}
	
	// Add custom certificates from trust store
	trustedCerts, err := r.trustStore.GetTrustedCerts(r.config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get trusted certificates from trust store: %v", err)
	}
	
	for _, cert := range trustedCerts {
		certPool.AddCert(cert)
	}
	
	tlsConfig := &tls.Config{
		ServerName:         r.baseURL.Host,
		RootCAs:           certPool,
		InsecureSkipVerify: false,
		MinVersion:        tls.VersionTLS12,
	}
	
	// Add certificate validation using Phase 1 validator
	tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("no certificates provided")
		}
		
		// Parse the leaf certificate
		leafCert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse leaf certificate: %v", err)
		}
		
		// Use Phase 1 validator to validate the certificate chain with hostname
		return r.validator.ValidateChainWithHostname(leafCert, r.baseURL.Host)
	}
	
	log.Printf("TLS configured with Phase 1 certificate infrastructure for %s", r.baseURL.Host)
	return tlsConfig, nil
}

// getBasicOptions returns minimal options when Phase 1 integration fails
func (r *giteaRegistryImpl) getBasicOptions() []remote.Option {
	var options []remote.Option
	
	// Basic authentication if available
	if r.authn != nil && r.authn.username != "" && r.authn.password != "" {
		auth := &remoteAuthenticator{
			username: r.authn.username,
			password: r.authn.password,
		}
		options = append(options, remote.WithAuth(auth))
	}
	
	// Basic transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: r.config.Insecure,
		},
	}
	options = append(options, remote.WithTransport(transport))
	
	return options
}

// remoteAuthenticator implements authn.Authenticator for go-containerregistry
type remoteAuthenticator struct {
	username string
	password string
	token    string
}

// Authorization returns the authentication header value
func (a *remoteAuthenticator) Authorization() (*authn.AuthConfig, error) {
	if a.token != "" {
		return &authn.AuthConfig{
			Auth: a.token,
		}, nil
	}
	
	if a.username != "" && a.password != "" {
		return &authn.AuthConfig{
			Username: a.username,
			Password: a.password,
		}, nil
	}
	
	return &authn.AuthConfig{}, nil
}