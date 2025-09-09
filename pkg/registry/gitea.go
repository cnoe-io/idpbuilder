package registry

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	
	// Phase 1 Certificate Infrastructure Integration
	"github.com/cnoe-io/idpbuilder/pkg/certs"
	"github.com/cnoe-io/idpbuilder/pkg/certvalidation"
	"github.com/cnoe-io/idpbuilder/pkg/fallback"
)

// giteaRegistryImpl implements the Registry interface for Gitea container registry.
// It integrates with Phase 1 certificate infrastructure for secure TLS connections
// and provides comprehensive error handling and retry logic.
type giteaRegistryImpl struct {
	config       RegistryConfig
	trustStore   certs.TrustStoreManager
	validator    certvalidation.CertValidator  
	fallback     fallback.FallbackHandler
	authn        *authenticator
	baseURL      *url.URL
	initialized  bool
}

// NewGiteaRegistry creates a new Gitea registry client with Phase 1 certificate integration.
// The registry config must include URL, username, and password at minimum.
// Returns error if configuration is invalid or Phase 1 components fail to initialize.
func NewGiteaRegistry(config RegistryConfig) (Registry, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("registry URL is required")
	}
	if config.Username == "" {
		return nil, fmt.Errorf("registry username is required")  
	}
	if config.Password == "" {
		return nil, fmt.Errorf("registry password is required")
	}
	
	// Parse and validate registry URL
	baseURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid registry URL %q: %v", config.URL, err)
	}
	
	// Set default timeout if not specified
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 30
	}
	
	// Initialize Phase 1 Certificate Infrastructure Components
	trustStore, err := certs.NewTrustStoreManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize trust store: %v", err)
	}
	
	validator, err := certvalidation.NewCertValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate validator: %v", err)
	}
	
	fallbackHandler, err := fallback.NewFallbackHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize fallback handler: %v", err)
	}
	
	// Create authenticator for credential management
	auth := &authenticator{
		username: config.Username,
		password: config.Password,
	}
	
	registry := &giteaRegistryImpl{
		config:      config,
		trustStore:  trustStore,
		validator:   validator,
		fallback:    fallbackHandler,
		authn:       auth,
		baseURL:     baseURL,
		initialized: true,
	}
	
	log.Printf("Gitea registry client initialized for %s", config.URL)
	return registry, nil
}

// validateRegistry ensures the registry instance is properly initialized
func (r *giteaRegistryImpl) validateRegistry() error {
	if !r.initialized {
		return fmt.Errorf("registry not properly initialized")
	}
	if r.trustStore == nil {
		return fmt.Errorf("trust store not initialized")
	}
	if r.validator == nil {
		return fmt.Errorf("certificate validator not initialized")
	}
	if r.fallback == nil {
		return fmt.Errorf("fallback handler not initialized") 
	}
	return nil
}

// buildRegistryURL constructs a full registry URL for operations
func (r *giteaRegistryImpl) buildRegistryURL(path string) string {
	return fmt.Sprintf("%s/%s", r.config.URL, path)
}

// getTimeout returns configured timeout as duration
func (r *giteaRegistryImpl) getTimeout() time.Duration {
	return time.Duration(r.config.TimeoutSeconds) * time.Second
}