package build

import "os"

const (
	// EnableImageBuilderFlag controls whether the OCI image builder is enabled
	EnableImageBuilderFlag = "ENABLE_IMAGE_BUILDER"
)

// IsImageBuilderEnabled checks if the OCI image builder feature is enabled
func IsImageBuilderEnabled() bool {
	value := os.Getenv(EnableImageBuilderFlag)
	return value == "true" || value == "1" || value == "enabled"
}
