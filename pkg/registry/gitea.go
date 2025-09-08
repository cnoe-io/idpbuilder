package registry

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// GiteaRegistry implements the Registry interface for Gitea container registry
type GiteaRegistry struct {
	baseURL    string
	httpClient *http.Client
	authMgr    *AuthManager
	config     *RemoteOptions
	logger     *logrus.Logger
}

// NewGiteaRegistry creates a new Gitea registry client
func NewGiteaRegistry(config *RegistryConfig, opts *RemoteOptions) (*GiteaRegistry, error) {
	if config == nil || config.URL == "" {
		return nil, fmt.Errorf("registry config and URL are required")
	}
	
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid registry URL: %w", err)
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}
	
	if opts == nil {
		opts = DefaultRemoteOptions()
	}
	
	// Create HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.Insecure},
	}
	if opts.ProxyURL != "" {
		if proxyURL, err := url.Parse(opts.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
	}
	
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	return &GiteaRegistry{
		baseURL:    parsedURL.String(),
		httpClient: httpClient,
		authMgr:    NewAuthManager(config.Username, config.Token),
		config:     opts,
		logger:     logger,
	}, nil
}

// Push uploads a container image to the registry
func (g *GiteaRegistry) Push(ctx context.Context, image string, content io.Reader) error {
	if image == "" || content == nil {
		return fmt.Errorf("image name and content are required")
	}
	
	g.logger.Infof("Pushing image: %s", image)
	
	pushURL := g.buildURL(fmt.Sprintf("v2/%s/manifests/latest", image))
	req, err := http.NewRequestWithContext(ctx, "PUT", pushURL, content)
	if err != nil {
		return fmt.Errorf("failed to create push request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("User-Agent", g.config.UserAgent)
	
	return g.executeWithRetry(ctx, req)
}

// List returns a list of repositories in the registry
func (g *GiteaRegistry) List(ctx context.Context) ([]string, error) {
	g.logger.Info("Listing repositories")
	
	catalogURL := g.buildURL("v2/_catalog")
	req, err := http.NewRequestWithContext(ctx, "GET", catalogURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}
	
	req.Header.Set("User-Agent", g.config.UserAgent)
	
	resp, err := g.executeRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return []string{}, nil // Catalog parsing would be implemented here
}

// Exists checks if a repository exists in the registry
func (g *GiteaRegistry) Exists(ctx context.Context, repository string) (bool, error) {
	if repository == "" {
		return false, fmt.Errorf("repository name is required")
	}
	
	manifestURL := g.buildURL(fmt.Sprintf("v2/%s/manifests/latest", repository))
	req, err := http.NewRequestWithContext(ctx, "HEAD", manifestURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create exists request: %w", err)
	}
	
	req.Header.Set("User-Agent", g.config.UserAgent)
	
	resp, err := g.executeRequest(ctx, req)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK, nil
}

// Delete removes a repository from the registry
func (g *GiteaRegistry) Delete(ctx context.Context, repository string) error {
	if repository == "" {
		return fmt.Errorf("repository name is required")
	}
	
	g.logger.Infof("Deleting repository: %s", repository)
	
	deleteURL := g.buildURL(fmt.Sprintf("v2/%s/manifests/latest", repository))
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	
	req.Header.Set("User-Agent", g.config.UserAgent)
	return g.executeWithRetry(ctx, req)
}

// Close cleans up any resources used by the registry client
func (g *GiteaRegistry) Close() error {
	g.logger.Info("Closing registry client")
	if transport, ok := g.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// buildURL constructs a full URL for the registry API
func (g *GiteaRegistry) buildURL(apiPath string) string {
	baseURL := strings.TrimSuffix(g.baseURL, "/")
	cleanPath := strings.TrimPrefix(apiPath, "/")
	return fmt.Sprintf("%s/%s", baseURL, cleanPath)
}

// executeRequest executes an HTTP request with authentication
func (g *GiteaRegistry) executeRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	authHeader, err := g.authMgr.GetAuthHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth header: %w", err)
	}
	
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	
	// Handle auth challenges
	if resp.StatusCode == http.StatusUnauthorized {
		wwwAuth := resp.Header.Get("WWW-Authenticate")
		if wwwAuth != "" {
			if err := g.authMgr.HandleAuthChallenge(wwwAuth); err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("failed to handle auth challenge: %w", err)
			}
			
			if err := g.authMgr.RefreshToken(ctx, g.httpClient); err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("failed to refresh token: %w", err)
			}
			
			resp.Body.Close()
			return g.executeRequest(ctx, req)
		}
	}
	
	return resp, nil
}

// executeWithRetry executes a request with retry logic
func (g *GiteaRegistry) executeWithRetry(ctx context.Context, req *http.Request) error {
	var lastErr error
	
	for attempt := 0; attempt <= g.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(g.config.RetryDelay):
			}
			g.logger.Warnf("Retrying request (attempt %d/%d)", attempt+1, g.config.MaxRetries+1)
		}
		
		reqCopy := req.Clone(ctx)
		resp, err := g.executeRequest(ctx, reqCopy)
		if err != nil {
			lastErr = err
			continue
		}
		
		defer resp.Body.Close()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lastErr = fmt.Errorf("request failed with status: %d", resp.StatusCode)
			continue
		}
		
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}
	
	return fmt.Errorf("request failed after %d attempts: %w", g.config.MaxRetries+1, lastErr)
}