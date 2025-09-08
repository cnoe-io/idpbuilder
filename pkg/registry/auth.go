package registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"
)

// authenticator handles authentication for registry operations.
// Manages credentials, token storage, and authentication state.
type authenticator struct {
	username      string
	password      string
	token         string
	authenticated bool
	lastAuth      time.Time
	authExpiry    time.Duration
}

// Authenticate performs registry authentication using configured credentials.
// Implements basic authentication with username/password and stores the auth token.
// Returns error if authentication fails or credentials are rejected.
func (r *giteaRegistryImpl) Authenticate(ctx context.Context) error {
	if err := r.validateRegistry(); err != nil {
		return fmt.Errorf("registry validation failed: %v", err)
	}
	
	if r.authn == nil {
		return fmt.Errorf("authenticator not initialized")
	}
	
	// Check if already authenticated and not expired
	if r.authn.authenticated && time.Since(r.authn.lastAuth) < r.authn.authExpiry {
		log.Printf("Using cached authentication for %s", r.config.Username)
		return nil
	}
	
	// Prepare basic authentication header
	authHeader := r.authn.createBasicAuthHeader()
	if authHeader == "" {
		return fmt.Errorf("failed to create authentication header")
	}
	
	// Create context with timeout
	authCtx, cancel := context.WithTimeout(ctx, r.getTimeout())
	defer cancel()
	
	// Perform authentication request
	if err := r.authn.performAuthentication(authCtx, r.buildRegistryURL("v2/"), authHeader); err != nil {
		r.authn.authenticated = false
		return fmt.Errorf("authentication failed for user %s: %v", r.config.Username, err)
	}
	
	// Mark as authenticated
	r.authn.authenticated = true
	r.authn.lastAuth = time.Now()
	r.authn.authExpiry = 15 * time.Minute // Token expires after 15 minutes
	
	log.Printf("Successfully authenticated user %s with registry", r.config.Username)
	return nil
}

// createBasicAuthHeader creates HTTP basic authentication header
func (a *authenticator) createBasicAuthHeader() string {
	if a.username == "" || a.password == "" {
		return ""
	}
	
	credentials := fmt.Sprintf("%s:%s", a.username, a.password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

// performAuthentication executes the authentication HTTP request
func (a *authenticator) performAuthentication(ctx context.Context, registryURL, authHeader string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", registryURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create authentication request: %v", err)
	}
	
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("User-Agent", "idpbuilder-oci/gitea-client")
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid credentials")
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %s", resp.Status)
	}
	
	// Store authentication token if provided
	if token := resp.Header.Get("Authorization"); token != "" {
		a.token = token
	}
	
	return nil
}

// IsAuthenticated returns true if the registry client is currently authenticated
func (a *authenticator) IsAuthenticated() bool {
	return a.authenticated && time.Since(a.lastAuth) < a.authExpiry
}

// GetAuthHeader returns the authentication header for registry requests
func (a *authenticator) GetAuthHeader() string {
	if a.token != "" {
		return a.token
	}
	return a.createBasicAuthHeader()
}