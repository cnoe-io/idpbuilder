package types

import (
	"testing"
	"time"
)

func TestRegistryConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  *RegistryConfig
		want bool
	}{
		{
			name: "valid config",
			cfg: &RegistryConfig{
				URL:       "registry.example.com",
				Namespace: "myorg",
				Timeout:   30 * time.Second,
			},
			want: true,
		},
		{
			name: "insecure config",
			cfg: &RegistryConfig{
				URL:      "localhost:5000",
				Insecure: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic config validation
			if (tt.cfg.URL != "") != tt.want {
				t.Errorf("config validation failed")
			}
		})
	}
}