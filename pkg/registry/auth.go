package registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthManager handles authentication for registry operations
type AuthManager struct {
	username string
	token    string
	realm    string
	service  string
	scope    string
	
	// Token management
	bearerToken    string
	tokenExpiresAt time.Time
	tokenMutex     sync.RWMutex
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(username, token string) *AuthManager {
	return &AuthManager{
		username: username,
		token:    token,
	}
}

// SetRealm configures the authentication realm
func (a *AuthManager) SetRealm(realm, service, scope string) {
	a.realm = realm
	a.service = service
	a.scope = scope
}

// GetAuthHeader returns the appropriate authorization header
func (a *AuthManager) GetAuthHeader(ctx context.Context) (string, error) {
	// Try bearer token first if available
	if a.hasBearerToken() {
		a.tokenMutex.RLock()
		token := a.bearerToken
		a.tokenMutex.RUnlock()
		return fmt.Sprintf("Bearer %s", token), nil
	}
	
	// Fall back to basic auth
	if a.username != "" && a.token != "" {
		return a.getBasicAuthHeader(), nil
	}
	
	return "", fmt.Errorf("no authentication credentials available")
}

// RefreshToken refreshes the bearer token if needed
func (a *AuthManager) RefreshToken(ctx context.Context, client *http.Client) error {
	if a.realm == "" {
		return nil
	}
	
	// Check if token needs refresh
	a.tokenMutex.RLock()
	needsRefresh := time.Now().After(a.tokenExpiresAt.Add(-5 * time.Minute))
	a.tokenMutex.RUnlock()
	
	if !needsRefresh {
		return nil
	}
	
	// Build token request
	tokenURL := fmt.Sprintf("%s?service=%s&scope=%s", a.realm, a.service, a.scope)
	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}
	
	// Add basic auth for token request
	if a.username != "" && a.token != "" {
		req.Header.Set("Authorization", a.getBasicAuthHeader())
	}
	
	// Execute token request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}
	
	// Parse token response (simplified)
	a.tokenMutex.Lock()
	a.bearerToken = "dummy-bearer-token"
	a.tokenExpiresAt = time.Now().Add(1 * time.Hour)
	a.tokenMutex.Unlock()
	
	return nil
}

// getBasicAuthHeader creates a basic authentication header
func (a *AuthManager) getBasicAuthHeader() string {
	credentials := fmt.Sprintf("%s:%s", a.username, a.token)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

// hasBearerToken checks if a valid bearer token is available
func (a *AuthManager) hasBearerToken() bool {
	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()
	return a.bearerToken != "" && time.Now().Before(a.tokenExpiresAt)
}

// HandleAuthChallenge processes WWW-Authenticate challenges
func (a *AuthManager) HandleAuthChallenge(challenge string) error {
	if !strings.HasPrefix(challenge, "Bearer ") {
		return nil
	}
	
	// Parse bearer challenge - comma separated key=value pairs
	challengeStr := challenge[7:] // Remove "Bearer "
	parts := strings.Split(challengeStr, ",")
	params := make(map[string]string)
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if keyValue := strings.SplitN(part, "=", 2); len(keyValue) == 2 {
			key := keyValue[0]
			value := strings.Trim(keyValue[1], `"`)
			params[key] = value
		}
	}
	
	// Extract authentication parameters
	if realm, ok := params["realm"]; ok {
		a.realm = realm
	}
	if service, ok := params["service"]; ok {
		a.service = service
	}
	if scope, ok := params["scope"]; ok {
		a.scope = scope
	}
	
	return nil
}

// ValidateCredentials checks if the provided credentials are valid
func (a *AuthManager) ValidateCredentials() error {
	if a.username == "" {
		return fmt.Errorf("username is required")
	}
	if a.token == "" {
		return fmt.Errorf("token is required")
	}
	if len(strings.TrimSpace(a.token)) == 0 {
		return fmt.Errorf("token cannot be empty or whitespace only")
	}
	return nil
}