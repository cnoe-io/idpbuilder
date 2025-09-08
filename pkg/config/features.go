package config

import (
	"os"
	"strconv"
	"strings"
)

// Feature flags for gradual activation of Gitea registry functionality
// These flags allow the registry client to be merged but remain disabled
// until the image-builder effort is also complete and tested.

// GITEA_REGISTRY_ENABLED controls whether Gitea registry operations are active
const GITEA_REGISTRY_ENABLED = "GITEA_REGISTRY_ENABLED"

// IsGiteaRegistryEnabled checks if Gitea registry operations are enabled.
// Returns false by default for safety until both E2.1.1 and E2.1.2 are complete.
func IsGiteaRegistryEnabled() bool {
	return getBoolEnv(GITEA_REGISTRY_ENABLED, false)
}

// getBoolEnv retrieves a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	// Handle common boolean representations
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "true", "1", "yes", "on", "enabled":
		return true
	case "false", "0", "no", "off", "disabled":
		return false
	default:
		// Try to parse as boolean
		if result, err := strconv.ParseBool(value); err == nil {
			return result
		}
		// Return default if parsing fails
		return defaultValue
	}
}