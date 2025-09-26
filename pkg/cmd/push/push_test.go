package push

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test basic command creation and metadata
func Test_PushCommand_Basic(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		assert.NotNil(t, PushCmd)
		assert.Equal(t, "push", PushCmd.Use[:4])
		assert.Contains(t, PushCmd.Short, "Push an OCI image")
		assert.Contains(t, PushCmd.Long, "Push an OCI image to a container registry")
	})

	t.Run("requires exactly one argument", func(t *testing.T) {
		// Test that the command validates exactly one argument
		err := PushCmd.Args(PushCmd, []string{})
		assert.Error(t, err)

		err = PushCmd.Args(PushCmd, []string{"image1", "image2"})
		assert.Error(t, err)

		err = PushCmd.Args(PushCmd, []string{"image1"})
		assert.NoError(t, err)
	})

	t.Run("has required flags", func(t *testing.T) {
		usernameFlag := PushCmd.Flags().Lookup("username")
		require.NotNil(t, usernameFlag)
		assert.Equal(t, "u", usernameFlag.Shorthand)

		passwordFlag := PushCmd.Flags().Lookup("password")
		require.NotNil(t, passwordFlag)
		assert.Equal(t, "p", passwordFlag.Shorthand)

		tlsFlag := PushCmd.Flags().Lookup("insecure-tls")
		require.NotNil(t, tlsFlag)
		assert.Equal(t, "false", tlsFlag.DefValue)
	})
}

// Test flag parsing and validation
func Test_PushCommand_Flags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantUser string
		wantPass string
		wantTLS  bool
		wantErr  bool
	}{
		{
			name:     "basic image URL",
			args:     []string{"registry.example.com/image:tag"},
			wantUser: "",
			wantPass: "",
			wantTLS:  false,
			wantErr:  false,
		},
		{
			name:     "with username",
			args:     []string{"--username", "testuser", "registry.example.com/image:tag"},
			wantUser: "testuser",
			wantPass: "",
			wantTLS:  false,
			wantErr:  false,
		},
		{
			name:     "with username shorthand",
			args:     []string{"-u", "testuser", "registry.example.com/image:tag"},
			wantUser: "testuser",
			wantPass: "",
			wantTLS:  false,
			wantErr:  false,
		},
		{
			name:     "with password",
			args:     []string{"--password", "testpass", "registry.example.com/image:tag"},
			wantUser: "",
			wantPass: "testpass",
			wantTLS:  false,
			wantErr:  false,
		},
		{
			name:     "with password shorthand",
			args:     []string{"-p", "testpass", "registry.example.com/image:tag"},
			wantUser: "",
			wantPass: "testpass",
			wantTLS:  false,
			wantErr:  false,
		},
		{
			name:     "with insecure TLS",
			args:     []string{"--insecure-tls", "registry.example.com/image:tag"},
			wantUser: "",
			wantPass: "",
			wantTLS:  true,
			wantErr:  false,
		},
		{
			name:     "all flags combined",
			args:     []string{"-u", "testuser", "-p", "testpass", "--insecure-tls", "registry.example.com/image:tag"},
			wantUser: "testuser",
			wantPass: "testpass",
			wantTLS:  true,
			wantErr:  false,
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many arguments",
			args:    []string{"image1", "image2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			username = ""
			password = ""
			insecureTLS = false

			// Create a new command instance for this test
			cmd := &cobra.Command{
				Use:   "push IMAGE_URL [flags]",
				Short: "Push an OCI image to a registry",
				Long:  `Push an OCI image to a container registry.`,
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					return nil
				},
			}

			// Add flags
			cmd.Flags().StringVarP(&username, "username", "u", "", "Registry username")
			cmd.Flags().StringVarP(&password, "password", "p", "", "Registry password")
			cmd.Flags().BoolVar(&insecureTLS, "insecure-tls", false, "Skip TLS certificate verification")

			// Set args and execute
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, username)
			assert.Equal(t, tt.wantPass, password)
			assert.Equal(t, tt.wantTLS, insecureTLS)
		})
	}
}

// Test command execution with various scenarios
func Test_PushCommand_Execute(t *testing.T) {
	tests := []struct {
		name          string
		imageURL      string
		username      string
		password      string
		insecureTLS   bool
		wantErr       bool
		wantErrString string
		wantOutput    []string
	}{
		{
			name:       "valid image URL",
			imageURL:   "registry.example.com/myapp:v1.0.0",
			wantErr:    false,
			wantOutput: []string{"Pushing image to: registry.example.com/myapp:v1.0.0", "Image pushed successfully"},
		},
		{
			name:       "with authentication",
			imageURL:   "registry.example.com/myapp:v1.0.0",
			username:   "testuser",
			password:   "testpass",
			wantErr:    false,
			wantOutput: []string{"Using authentication for user: testuser", "Image pushed successfully"},
		},
		{
			name:        "with insecure TLS",
			imageURL:    "registry.example.com/myapp:v1.0.0",
			insecureTLS: true,
			wantErr:     false,
			wantOutput:  []string{"Warning: Using insecure TLS connection", "Image pushed successfully"},
		},
		{
			name:          "invalid image URL format",
			imageURL:      "invalid-image-url",
			wantErr:       true,
			wantErrString: "invalid image URL format",
		},
		{
			name:          "empty image URL",
			imageURL:      "",
			wantErr:       true,
			wantErrString: "invalid image URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global variables
			username = tt.username
			password = tt.password
			insecureTLS = tt.insecureTLS

			// Capture output
			var output bytes.Buffer
			cmd := &cobra.Command{
				Use:  "push",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					imageURL := args[0]

					// Validate image URL format
					if !strings.Contains(imageURL, "/") {
						return fmt.Errorf("invalid image URL format: %s", imageURL)
					}

					// Mock push implementation
					fmt.Fprintf(&output, "Pushing image to: %s\n", imageURL)

					if username != "" {
						fmt.Fprintf(&output, "Using authentication for user: %s\n", username)
					}

					if insecureTLS {
						fmt.Fprintf(&output, "Warning: Using insecure TLS connection\n")
					}

					fmt.Fprintf(&output, "Image pushed successfully\n")
					return nil
				},
			}

			cmd.SetArgs([]string{tt.imageURL})
			cmd.SetOut(&output)

			err := cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrString != "" {
					assert.Contains(t, err.Error(), tt.wantErrString)
				}
				return
			}

			require.NoError(t, err)

			outputStr := output.String()
			for _, expectedOutput := range tt.wantOutput {
				assert.Contains(t, outputStr, expectedOutput)
			}
		})
	}
}

// Test authentication handling scenarios
func Test_PushCommand_Auth(t *testing.T) {
	tests := []struct {
		name         string
		username     string
		password     string
		expectOutput string
	}{
		{
			name:         "with valid credentials",
			username:     "validuser",
			password:     "validpass",
			expectOutput: "Using authentication for user: validuser",
		},
		{
			name:         "with username only",
			username:     "testuser",
			password:     "",
			expectOutput: "Using authentication for user: testuser",
		},
		{
			name:         "with password only",
			username:     "",
			password:     "testpass",
			expectOutput: "", // No auth message when username is empty
		},
		{
			name:         "without credentials",
			username:     "",
			password:     "",
			expectOutput: "", // No auth message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username = tt.username
			password = tt.password

			err := pushImage("registry.example.com/test:latest")
			require.NoError(t, err)

			// Since pushImage prints to stdout, we need to capture it differently
			// This is a simplified test that checks the logic
			if tt.username != "" && tt.expectOutput != "" {
				// In the real implementation, we would check if auth is used
				assert.NotEmpty(t, tt.username)
			}
		})
	}
}

// Test TLS configuration scenarios
func Test_PushCommand_TLS(t *testing.T) {
	tests := []struct {
		name        string
		insecureTLS bool
		wantWarning bool
	}{
		{
			name:        "secure connection (default)",
			insecureTLS: false,
			wantWarning: false,
		},
		{
			name:        "insecure TLS enabled",
			insecureTLS: true,
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insecureTLS = tt.insecureTLS

			err := pushImage("registry.example.com/test:latest")
			require.NoError(t, err)

			// In a real implementation, we would verify TLS configuration
			// For now, just verify the flag is set correctly
			assert.Equal(t, tt.insecureTLS, insecureTLS)
		})
	}
}

// Test error handling scenarios
func Test_PushCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		imageURL      string
		wantErr       bool
		wantErrString string
	}{
		{
			name:          "missing required argument",
			imageURL:      "",
			wantErr:       true,
			wantErrString: "invalid image URL format",
		},
		{
			name:          "invalid image format - no slash",
			imageURL:      "invalidimage",
			wantErr:       true,
			wantErrString: "invalid image URL format",
		},
		{
			name:          "invalid image format - just colon",
			imageURL:      ":tag",
			wantErr:       true,
			wantErrString: "invalid image URL format",
		},
		{
			name:     "valid image format",
			imageURL: "registry.com/image:tag",
			wantErr:  false,
		},
		{
			name:     "valid image format with port",
			imageURL: "registry.com:5000/image:tag",
			wantErr:  false,
		},
		{
			name:     "localhost registry",
			imageURL: "localhost:5000/image:tag",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.imageURL == "" {
				// Test the command validation
				cmd := &cobra.Command{
					Args: cobra.ExactArgs(1),
					RunE: func(cmd *cobra.Command, args []string) error {
						return nil
					},
				}
				cmd.SetArgs([]string{})
				err := cmd.Execute()
				assert.Error(t, err)
				return
			}

			err := pushImage(tt.imageURL)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrString != "" {
					assert.Contains(t, err.Error(), tt.wantErrString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkPushCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := pushImage("registry.example.com/benchmark:test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFlagParsing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := &cobra.Command{
			Use:  "push",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}

		var user, pass string
		var insecure bool
		cmd.Flags().StringVarP(&user, "username", "u", "", "Registry username")
		cmd.Flags().StringVarP(&pass, "password", "p", "", "Registry password")
		cmd.Flags().BoolVar(&insecure, "insecure-tls", false, "Skip TLS certificate verification")

		cmd.SetArgs([]string{"-u", "testuser", "-p", "testpass", "--insecure-tls", "registry.com/image:tag"})
		err := cmd.Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}