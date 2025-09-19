package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewGiteaRegistry(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &RegistryConfig{
			URL:      "https://gitea.example.com",
			Username: "testuser",
			Token:    "testtoken",
		}
		
		registry, err := NewGiteaRegistry(config, nil)
		if err != nil {
			t.Fatalf("NewGiteaRegistry() error = %v", err)
		}
		
		if registry == nil {
			t.Fatal("Expected registry to be created")
		}
		
		if registry.baseURL != "https://gitea.example.com" {
			t.Errorf("Expected baseURL https://gitea.example.com, got %s", registry.baseURL)
		}
	})
	
	t.Run("nil config", func(t *testing.T) {
		_, err := NewGiteaRegistry(nil, nil)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})
	
	t.Run("empty URL", func(t *testing.T) {
		config := &RegistryConfig{}
		_, err := NewGiteaRegistry(config, nil)
		if err == nil {
			t.Error("Expected error for empty URL")
		}
	})
	
	t.Run("invalid URL", func(t *testing.T) {
		config := &RegistryConfig{
			URL: "://invalid-url",
		}
		_, err := NewGiteaRegistry(config, nil)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})
}

func TestGiteaRegistry_buildURL(t *testing.T) {
	config := &RegistryConfig{
		URL:      "https://gitea.example.com/",
		Username: "testuser",
		Token:    "testtoken",
	}
	
	registry, err := NewGiteaRegistry(config, nil)
	if err != nil {
		t.Fatalf("NewGiteaRegistry() error = %v", err)
	}
	
	tests := []struct {
		name     string
		apiPath  string
		expected string
	}{
		{
			name:     "catalog path",
			apiPath:  "v2/_catalog",
			expected: "https://gitea.example.com/v2/_catalog",
		},
		{
			name:     "manifest path",
			apiPath:  "/v2/repo/manifests/latest",
			expected: "https://gitea.example.com/v2/repo/manifests/latest",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.buildURL(tt.apiPath)
			if result != tt.expected {
				t.Errorf("buildURL() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGiteaRegistry_Close(t *testing.T) {
	config := &RegistryConfig{
		URL:      "https://gitea.example.com",
		Username: "testuser",
		Token:    "testtoken",
	}
	
	registry, err := NewGiteaRegistry(config, nil)
	if err != nil {
		t.Fatalf("NewGiteaRegistry() error = %v", err)
	}
	
	err = registry.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestGiteaRegistry_Exists(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		if strings.Contains(r.URL.Path, "exists") {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	config := &RegistryConfig{
		URL:      server.URL,
		Username: "testuser",
		Token:    "testtoken",
	}
	
	registry, err := NewGiteaRegistry(config, nil)
	if err != nil {
		t.Fatalf("NewGiteaRegistry() error = %v", err)
	}
	defer registry.Close()
	
	ctx := context.Background()
	
	t.Run("repository exists", func(t *testing.T) {
		exists, err := registry.Exists(ctx, "exists")
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Expected repository to exist")
		}
	})
	
	t.Run("repository does not exist", func(t *testing.T) {
		exists, err := registry.Exists(ctx, "notfound")
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if exists {
			t.Error("Expected repository to not exist")
		}
	})
	
	t.Run("empty repository name", func(t *testing.T) {
		_, err := registry.Exists(ctx, "")
		if err == nil {
			t.Error("Expected error for empty repository name")
		}
	})
}

func TestDefaultRemoteOptions(t *testing.T) {
	opts := DefaultRemoteOptions()
	
	if opts == nil {
		t.Fatal("Expected default options to be created")
	}
	
	if opts.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", opts.Timeout)
	}
	
	if opts.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", opts.MaxRetries)
	}
	
	if opts.UserAgent != "idpbuilder-gitea-client/1.0" {
		t.Errorf("Expected user agent idpbuilder-gitea-client/1.0, got %s", opts.UserAgent)
	}
}