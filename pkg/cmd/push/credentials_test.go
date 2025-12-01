package push

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEnvironment implements EnvironmentLookup for testing.
type MockEnvironment struct {
	mock.Mock
}

// Get implements EnvironmentLookup.Get for mocking.
func (m *MockEnvironment) Get(key string) string {
	args := m.Called(key)
	return args.String(0)
}

// TestCredentialResolver_FlagPrecedence tests the complete credential resolution
// logic with table-driven tests covering all resolution scenarios.
func TestCredentialResolver_FlagPrecedence(t *testing.T) {
	tests := []struct {
		name          string
		flags         CredentialFlags
		envUsername   string
		envPassword   string
		envToken      string
		wantUsername  string
		wantPassword  string
		wantToken     string
		wantAnonymous bool
		wantErr       bool
	}{
		// Test Case 1: Flag overrides environment for username/password
		{
			name:         "flag_overrides_env_username",
			flags:        CredentialFlags{Username: "flag-user", Password: "flag-pass"},
			envUsername:  "env-user",
			envPassword:  "env-pass",
			wantUsername: "flag-user",
			wantPassword: "flag-pass",
			wantAnonymous: false,
		},

		// Test Case 2: Environment used when no flags provided
		{
			name:         "env_used_when_no_flags",
			flags:        CredentialFlags{},
			envUsername:  "env-user",
			envPassword:  "env-pass",
			wantUsername: "env-user",
			wantPassword: "env-pass",
			wantAnonymous: false,
		},

		// Test Case 3: Token flag overrides token env
		{
			name:      "token_flag_overrides_token_env",
			flags:     CredentialFlags{Token: "flag-token"},
			envToken:  "env-token",
			wantToken: "flag-token",
			wantAnonymous: false,
		},

		// Test Case 4: Token env used when no token flag
		{
			name:      "token_env_used_when_no_token_flag",
			flags:     CredentialFlags{},
			envToken:  "env-token",
			wantToken: "env-token",
			wantAnonymous: false,
		},

		// Test Case 5: Anonymous access when no credentials
		{
			name:          "anonymous_when_no_credentials",
			flags:         CredentialFlags{},
			wantAnonymous: true,
		},

		// Test Case 6: Error when both token and basic auth
		{
			name:    "error_when_both_token_and_basic_auth",
			flags:   CredentialFlags{Username: "user", Token: "token"},
			wantErr: true,
		},

		// Test Case 7: Partial flag override (flag username, env password)
		{
			name:         "partial_flag_override",
			flags:        CredentialFlags{Username: "flag-user"},
			envUsername:  "env-user",
			envPassword:  "env-pass",
			wantUsername: "flag-user",
			wantPassword: "env-pass", // From environment
			wantAnonymous: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock environment
			// Note: Resolver only calls env.Get() when the corresponding flag is empty,
			// so we conditionally set up expectations based on flag values.
			mockEnv := new(MockEnvironment)

			// Token env lookup happens only if flag is empty
			if tt.flags.Token == "" {
				mockEnv.On("Get", EnvRegistryToken).Return(tt.envToken)
			}

			// Username env lookup happens only if flag is empty
			if tt.flags.Username == "" {
				mockEnv.On("Get", EnvRegistryUsername).Return(tt.envUsername)
			}

			// Password env lookup happens only if flag is empty
			if tt.flags.Password == "" {
				mockEnv.On("Get", EnvRegistryPassword).Return(tt.envPassword)
			}

			// Execute resolution
			resolver := &DefaultCredentialResolver{}
			creds, err := resolver.Resolve(tt.flags, mockEnv)

			// Validate error cases
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				assert.Contains(t, err.Error(), "cannot specify both token and username/password")
				// Mock expectations still apply for error cases
				mockEnv.AssertExpectations(t)
				return
			}

			// Validate success cases
			require.NoError(t, err, "Unexpected error for test case: %s", tt.name)
			require.NotNil(t, creds, "Credentials should not be nil")

			// Validate credential values
			assert.Equal(t, tt.wantUsername, creds.Username, "Username mismatch")
			assert.Equal(t, tt.wantPassword, creds.Password, "Password mismatch")
			assert.Equal(t, tt.wantToken, creds.Token, "Token mismatch")
			assert.Equal(t, tt.wantAnonymous, creds.IsAnonymous, "IsAnonymous mismatch")

			// Verify mock expectations
			mockEnv.AssertExpectations(t)
		})
	}
}

// TestCredentialResolver_NoCredentialLogging verifies that the Credentials struct
// does NOT have a String() method that could accidentally expose secrets in logs.
// This implements security property P1.3 from the phase architecture.
func TestCredentialResolver_NoCredentialLogging(t *testing.T) {
	creds := &Credentials{
		Username: "secret-user",
		Password: "secret-pass",
		Token:    "secret-token",
	}

	// Verify struct exists with expected fields
	assert.NotEmpty(t, creds.Username, "Username field should be populated")
	assert.NotEmpty(t, creds.Password, "Password field should be populated")
	assert.NotEmpty(t, creds.Token, "Token field should be populated")

	// Note: Credentials struct intentionally has NO String() method
	// to prevent accidental credential logging. This test documents
	// that security requirement and would fail compilation if String()
	// were added (due to type assertion below).

	// Verify Credentials does not implement fmt.Stringer interface
	_, implementsStringer := interface{}(creds).(interface{ String() string })
	assert.False(t, implementsStringer, "Credentials MUST NOT implement String() method for security")
}

// TestDefaultEnvironment_Get verifies that DefaultEnvironment correctly
// wraps os.Getenv for production use.
func TestDefaultEnvironment_Get(t *testing.T) {
	env := &DefaultEnvironment{}

	// Test with a known environment variable (PATH should always exist)
	value := env.Get("PATH")
	assert.NotEmpty(t, value, "PATH environment variable should exist")

	// Test with non-existent variable
	value = env.Get("IDPBUILDER_NONEXISTENT_VAR_12345")
	assert.Empty(t, value, "Non-existent variable should return empty string")
}
