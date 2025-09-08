package build

import (
	"github.com/google/go-containerregistry/pkg/v1"
)

// BuildOptions contains configuration for building OCI images
type BuildOptions struct {
	// ContextPath is the directory path to build from
	ContextPath string
	// Tag is the name and tag for the built image
	Tag string
	// Exclusions are patterns to exclude from the build context (like .dockerignore)
	Exclusions []string
	// Labels are key-value pairs to add to the image metadata
	Labels map[string]string
}

// BuildResult contains the result of an image build operation
type BuildResult struct {
	// ImageID is the unique identifier of the built image
	ImageID string
	// Digest is the content hash of the built image
	Digest v1.Hash
	// Size is the total size of the built image in bytes
	Size int64
	// StoragePath is the local path where the image was saved
	StoragePath string
}

// Builder is responsible for building OCI images from directories
type Builder struct {
	// storageDir is the directory where built images are stored locally
	storageDir string
	// images maps image tags to their local storage paths
	images map[string]string
}