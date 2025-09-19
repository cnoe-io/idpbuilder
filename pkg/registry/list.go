package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// RepositoryInfo represents detailed information about a repository
type RepositoryInfo struct {
	Name     string    `json:"name"`
	FullName string    `json:"full_name,omitempty"`
	Tags     []TagInfo `json:"tags,omitempty"`
}

// TagInfo represents information about a repository tag
type TagInfo struct {
	Name   string `json:"name"`
	Digest string `json:"digest,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

// ListOptions provides pagination and filtering options for listing operations
type ListOptions struct {
	Page        int  `json:"page,omitempty"`
	PageSize    int  `json:"page_size,omitempty"`
	IncludeTags bool `json:"include_tags,omitempty"`
}

// DefaultListOptions returns default pagination settings
func DefaultListOptions() *ListOptions {
	return &ListOptions{Page: 1, PageSize: 50}
}

// RepositoryLister provides enhanced repository listing capabilities
type RepositoryLister struct {
	registry *GiteaRegistry
}

// NewRepositoryLister creates a new repository lister instance
func NewRepositoryLister(registry *GiteaRegistry) *RepositoryLister {
	return &RepositoryLister{registry}
}

// ListRepositories returns a list of repositories with optional tag information
func (l *RepositoryLister) ListRepositories(ctx context.Context, opts *ListOptions) ([]RepositoryInfo, error) {
	if opts == nil { opts = DefaultListOptions() }

	catalogURL := fmt.Sprintf("%s/v2/_catalog", l.registry.baseURL)
	if opts.PageSize > 0 { catalogURL += "?n=" + strconv.Itoa(opts.PageSize) }

	req, err := http.NewRequestWithContext(ctx, "GET", catalogURL, nil)
	if err != nil { return nil, err }
	if authHeader, err := l.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := l.registry.httpClient.Do(req)
	if err != nil { return nil, l.registry.handleListingError(err, "request execution") }
	defer resp.Body.Close()

	if err := l.registry.validateCatalogResponse(resp); err != nil { return nil, err }

	var catalogResponse struct { Repositories []string `json:"repositories"` }
	if err := json.NewDecoder(resp.Body).Decode(&catalogResponse); err != nil { return nil, err }

	repositories := make([]RepositoryInfo, len(catalogResponse.Repositories))
	for i, repoName := range catalogResponse.Repositories {
		repo := RepositoryInfo{Name: repoName, FullName: repoName}
		if opts.IncludeTags {
			if tags, err := l.ListTags(ctx, repoName, opts); err == nil { repo.Tags = tags }
		}
		repositories[i] = repo
	}
	return repositories, nil
}

// ListTags returns tags for a specific repository
func (l *RepositoryLister) ListTags(ctx context.Context, repository string, opts *ListOptions) ([]TagInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v2/%s/tags/list", l.registry.baseURL, repository), nil)
	if err != nil { return nil, err }
	if authHeader, err := l.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := l.registry.httpClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound { return []TagInfo{}, nil }
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode) }

	var tagsResponse struct { Tags []string `json:"tags"` }
	if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil { return nil, err }

	tags := make([]TagInfo, len(tagsResponse.Tags))
	for i, tagName := range tagsResponse.Tags { tags[i] = TagInfo{Name: tagName} }
	return tags, nil
}

// SearchRepositories filters repositories by query string
func (l *RepositoryLister) SearchRepositories(ctx context.Context, query string, opts *ListOptions) ([]RepositoryInfo, error) {
	repos, err := l.ListRepositories(ctx, opts)
	if err != nil { return nil, err }

	var filtered []RepositoryInfo
	for _, repo := range repos {
		if strings.Contains(repo.Name, query) { filtered = append(filtered, repo) }
	}
	return filtered, nil
}

// ListRepositories returns a list of available repository names in the registry.
// This is the basic implementation that integrates with the enhanced retry logic from split-001.
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