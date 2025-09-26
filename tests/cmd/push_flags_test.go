package cmd_test

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/auth"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/push"
	"github.com/spf13/cobra"
)

func TestPushCommandFlags(t *testing.T) {
	// Create a new push command for testing
	cmd := push.PushCmd

	// Test that username flag exists
	usernameFlag := cmd.Flags().Lookup("username")
	if usernameFlag == nil {
		t.Error("Expected username flag to exist")
	}

	// Test that password flag exists
	passwordFlag := cmd.Flags().Lookup("password")
	if passwordFlag == nil {
		t.Error("Expected password flag to exist")
	}

	// Test short flag aliases
	usernameFlagShort := cmd.Flags().Lookup("u")
	if usernameFlagShort == nil {
		t.Error("Expected username short flag (-u) to exist")
	}

	passwordFlagShort := cmd.Flags().Lookup("p")
	if passwordFlagShort == nil {
		t.Error("Expected password short flag (-p) to exist")
	}
}

func TestCredentialExtraction(t *testing.T) {
	cmd := &cobra.Command{}
	auth.AddAuthenticationFlags(cmd)

	// Set test values
	cmd.Flags().Set("username", "testuser")
	cmd.Flags().Set("password", "testpass")

	// Extract credentials
	creds, err := auth.ExtractCredentialsFromFlags(cmd)
	if err != nil {
		t.Fatalf("Failed to extract credentials: %v", err)
	}

	// Verify extracted values
	if creds.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", creds.Username)
	}

	if creds.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", creds.Password)
	}
}