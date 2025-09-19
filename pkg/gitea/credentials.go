package gitea

import (
	"fmt"
	"os"
)

// CredentialProvider defines interface for credential sources
type CredentialProvider interface {
	GetUsername() (string, error)
	GetPassword() (string, error)
	IsAvailable() bool
	Priority() int
}

// CredentialManager manages multiple credential providers
type CredentialManager struct {
	providers []CredentialProvider
}

// NewCredentialManager creates a credential manager with default providers
func NewCredentialManager() *CredentialManager {
	return &CredentialManager{
		providers: []CredentialProvider{
			NewCLICredentialProvider(),
			NewEnvCredentialProvider(),
			NewConfigFileProvider(),
			NewKeyringProvider(),
		},
	}
}

// SetCLICredentials sets credentials provided via CLI flags
func (cm *CredentialManager) SetCLICredentials(username, password string) {
	if cliProvider, ok := cm.providers[0].(*CLICredentialProvider); ok {
		cliProvider.SetCredentials(username, password)
	}
}

// GetCredentials retrieves credentials from the first available provider
func (cm *CredentialManager) GetCredentials() (username, password string, err error) {
	for _, provider := range cm.providers {
		if provider.IsAvailable() {
			username, err = provider.GetUsername()
			if err != nil {
				continue
			}
			password, err = provider.GetPassword()
			if err != nil {
				continue
			}
			return username, password, nil
		}
	}
	return "", "", fmt.Errorf("no credentials available from any provider")
}

// GetUsername retrieves username from the first available provider (backward compatibility)
func (cm *CredentialManager) GetUsername() string {
	username, _, err := cm.GetCredentials()
	if err != nil {
		return ""
	}
	return username
}

// GetPassword retrieves password from the first available provider (backward compatibility)
func (cm *CredentialManager) GetPassword() string {
	_, password, err := cm.GetCredentials()
	if err != nil {
		return ""
	}
	return password
}

// EnvCredentialProvider reads from environment variables
type EnvCredentialProvider struct{}

func NewEnvCredentialProvider() *EnvCredentialProvider {
	return &EnvCredentialProvider{}
}

func (e *EnvCredentialProvider) GetUsername() (string, error) {
	username := os.Getenv("GITEA_USERNAME")
	if username == "" {
		return "", fmt.Errorf("GITEA_USERNAME not set")
	}
	return username, nil
}

func (e *EnvCredentialProvider) GetPassword() (string, error) {
	password := os.Getenv("GITEA_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("GITEA_PASSWORD not set")
	}
	return password, nil
}

func (e *EnvCredentialProvider) IsAvailable() bool {
	return os.Getenv("GITEA_USERNAME") != "" && os.Getenv("GITEA_PASSWORD") != ""
}

func (e *EnvCredentialProvider) Priority() int {
	return 2 // Second priority after CLI
}

// CLICredentialProvider holds credentials from command-line flags
type CLICredentialProvider struct {
	username string
	password string
}

func NewCLICredentialProvider() *CLICredentialProvider {
	return &CLICredentialProvider{}
}

func (c *CLICredentialProvider) SetCredentials(username, password string) {
	c.username = username
	c.password = password
}

func (c *CLICredentialProvider) GetUsername() (string, error) {
	if c.username == "" {
		return "", fmt.Errorf("no CLI username provided")
	}
	return c.username, nil
}

func (c *CLICredentialProvider) GetPassword() (string, error) {
	if c.password == "" {
		return "", fmt.Errorf("no CLI password/token provided")
	}
	return c.password, nil
}

func (c *CLICredentialProvider) IsAvailable() bool {
	return c.username != "" && c.password != ""
}

func (c *CLICredentialProvider) Priority() int {
	return 1 // Highest priority
}