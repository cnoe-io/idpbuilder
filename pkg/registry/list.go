package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// ListRepositories returns a list of available repository names in the registry.
// Requires authentication and uses the Docker Registry v2 API catalog endpoint.
// Returns empty slice if no repositories found or user lacks permissions.
func (r *GiteaRegistry) ListRepositories(ctx context.Context) ([]string, error) {
	if r.baseURL == "" {
		return nil, fmt.Errorf("registry baseURL is required")
	}

	r.logger.Printf("Listing repositories from registry %s", r.baseURL)
	
	// Perform repository listing with retry
	operation := func() ([]string, error) {
		return r.executeRepositoryListing(ctx)
	}
	
	return r.retryRepositoryOperation(operation, "list repositories")
}

// executeRepositoryListing performs the actual API call to list repositories
func (r *GiteaRegistry) executeRepositoryListing(ctx context.Context) ([]string, error) {
	// Build catalog endpoint URL
	catalogURL := r.buildCatalogURL()
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", catalogURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create catalog request: %v", err)
	}
	
	// Add authentication headers
	if authHeader, err := r.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	req.Header.Set("User-Agent", "idpbuilder-oci/gitea-client")
	req.Header.Set("Accept", "application/json")
	
	// Use the existing HTTP client from GiteaRegistry
	client := r.httpClient
	
	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, r.handleListingError(err, "request execution")
	}
	defer resp.Body.Close()
	
	// Handle HTTP response
	if err := r.validateCatalogResponse(resp); err != nil {
		return nil, err
	}
	
	// Parse the response
	repositories, err := r.parseCatalogResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse catalog response: %v", err)
	}
	
	log.Printf("Found %d repositories in registry", len(repositories))
	return repositories, nil
}

// buildCatalogURL constructs the Docker Registry v2 catalog endpoint URL
func (r *GiteaRegistry) buildCatalogURL() string {
	return fmt.Sprintf("%s/v2/_catalog", strings.TrimSuffix(r.baseURL, "/"))
}

// This method is no longer needed since we use the existing httpClient

// validateCatalogResponse validates the HTTP response from catalog endpoint
func (r *GiteaRegistry) validateCatalogResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
		
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: authentication required or invalid credentials")
		
	case http.StatusForbidden:
		return fmt.Errorf("forbidden: insufficient permissions to list repositories")
		
	case http.StatusNotFound:
		return fmt.Errorf("catalog endpoint not found: registry may not support v2 API")
		
	case http.StatusInternalServerError:
		return fmt.Errorf("registry internal error: %s", resp.Status)
		
	default:
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}
}

// catalogResponse represents the JSON structure returned by the catalog endpoint
type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

// parseCatalogResponse parses the JSON response from catalog endpoint
func (r *GiteaRegistry) parseCatalogResponse(body io.Reader) ([]string, error) {
	var response catalogResponse
	
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %v", err)
	}
	
	// Filter out any empty or invalid repository names
	validRepos := make([]string, 0, len(response.Repositories))
	for _, repo := range response.Repositories {
		if strings.TrimSpace(repo) != "" {
			validRepos = append(validRepos, strings.TrimSpace(repo))
		}
	}
	
	return validRepos, nil
}

// handleListingError provides comprehensive error handling for listing failures
func (r *GiteaRegistry) handleListingError(err error, context string) error {
	errorMsg := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errorMsg, "timeout"):
		return fmt.Errorf("repository listing timed out: %v", err)
		
	case strings.Contains(errorMsg, "connection refused"):
		return fmt.Errorf("connection refused: registry may be unavailable")
		
	case strings.Contains(errorMsg, "tls"):
		if r.config.Insecure {
			return fmt.Errorf("TLS error despite insecure mode: %v", err)
		}
		return fmt.Errorf("TLS certificate error: %v (try --insecure for development)", err)
		
	case strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "connection"):
		return fmt.Errorf("network error during repository listing: %v", err)
		
	default:
		return fmt.Errorf("repository listing failed during %s: %v", context, err)
	}
}

// retryRepositoryOperation performs retry logic for repository operations
func (r *GiteaRegistry) retryRepositoryOperation(operation func() ([]string, error), operationName string) ([]string, error) {
	var lastErr error
	var result []string
	
	retryOperation := func() error {
		var err error
		result, err = operation()
		lastErr = err
		return err
	}
	
	if err := retryWithExponentialBackoff(retryOperation, operationName, r.baseURL); err != nil {
		return nil, lastErr
	}
	
	return result, nil
}