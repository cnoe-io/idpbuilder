package registry

import (
	"context"
	"strings"
	"testing"
)

func TestNewAuthManager(t *testing.T) {
	username := "testuser"
	token := "testtoken"
	
	auth := NewAuthManager(username, token)
	
	if auth.username != username {
		t.Errorf("Expected username %s, got %s", username, auth.username)
	}
	
	if auth.token != token {
		t.Errorf("Expected token %s, got %s", token, auth.token)
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		username string
		token    string
		wantErr  bool
	}{
		{
			name:     "valid credentials",
			username: "testuser",
			token:    "testtoken",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			token:    "testtoken",
			wantErr:  true,
		},
		{
			name:     "empty token",
			username: "testuser",
			token:    "",
			wantErr:  true,
		},
		{
			name:     "whitespace token",
			username: "testuser",
			token:    "   ",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAuthManager(tt.username, tt.token)
			err := auth.ValidateCredentials()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAuthHeader(t *testing.T) {
	ctx := context.Background()
	
	t.Run("basic auth", func(t *testing.T) {
		auth := NewAuthManager("testuser", "testtoken")
		
		header, err := auth.GetAuthHeader(ctx)
		if err != nil {
			t.Fatalf("GetAuthHeader() error = %v", err)
		}
		
		if !strings.HasPrefix(header, "Basic ") {
			t.Errorf("Expected Basic auth header, got %s", header)
		}
	})
	
	t.Run("no credentials", func(t *testing.T) {
		auth := NewAuthManager("", "")
		
		_, err := auth.GetAuthHeader(ctx)
		if err == nil {
			t.Error("Expected error for no credentials")
		}
	})
}

func TestSetRealm(t *testing.T) {
	auth := NewAuthManager("user", "token")
	realm := "https://registry.example.com/auth"
	service := "registry"
	scope := "repository:test:pull,push"
	
	auth.SetRealm(realm, service, scope)
	
	if auth.realm != realm {
		t.Errorf("Expected realm %s, got %s", realm, auth.realm)
	}
	if auth.service != service {
		t.Errorf("Expected service %s, got %s", service, auth.service)
	}
	if auth.scope != scope {
		t.Errorf("Expected scope %s, got %s", scope, auth.scope)
	}
}

func TestHandleAuthChallenge(t *testing.T) {
	auth := NewAuthManager("user", "token")
	
	challenge := `Bearer realm="https://registry.example.com/auth",service="registry",scope="repository:test:pull"`
	
	err := auth.HandleAuthChallenge(challenge)
	if err != nil {
		t.Fatalf("HandleAuthChallenge() error = %v", err)
	}
	
	if auth.realm != "https://registry.example.com/auth" {
		t.Errorf("Expected realm to be parsed correctly, got %s", auth.realm)
	}
	if auth.service != "registry" {
		t.Errorf("Expected service to be parsed correctly, got %s", auth.service)
	}
	if auth.scope != "repository:test:pull" {
		t.Errorf("Expected scope to be parsed correctly, got %s", auth.scope)
	}
}