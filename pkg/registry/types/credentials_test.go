package types

import "testing"

func TestAuthConfig(t *testing.T) {
	tests := []struct {
		name    string
		auth    *AuthConfig
		wantErr bool
	}{
		{
			name: "basic auth",
			auth: &AuthConfig{
				AuthType: AuthTypeBasic,
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "token auth",
			auth: &AuthConfig{
				AuthType: AuthTypeToken,
				Token:    "bearer-token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.auth == nil && !tt.wantErr {
				t.Error("expected non-nil auth")
			}
		})
	}
}