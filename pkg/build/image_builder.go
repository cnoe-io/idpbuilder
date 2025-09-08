package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

var (
	// ErrFeatureDisabled is returned when the image builder feature is disabled
	ErrFeatureDisabled = errors.New("image builder feature is disabled")
)

// NewBuilder creates a new OCI image builder with the specified storage directory
func NewBuilder(storageDir string) (*Builder, error) {
	if storageDir == "" {
		return nil, fmt.Errorf("storage directory cannot be empty")
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Builder{
		storageDir: storageDir,
		images:     make(map[string]string),
	}, nil
}

// BuildImage builds an OCI image from a directory context
func (b *Builder) BuildImage(ctx context.Context, opts BuildOptions) (*BuildResult, error) {
	// Check feature flag
	if !IsImageBuilderEnabled() {
		return nil, ErrFeatureDisabled
	}

	// Validate options
	if opts.ContextPath == "" {
		return nil, fmt.Errorf("context path cannot be empty")
	}
	if opts.Tag == "" {
		return nil, fmt.Errorf("tag cannot be empty")
	}

	// Create tar archive from context directory
	tarReader, err := createTarFromContext(opts.ContextPath, opts.Exclusions)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar from context: %w", err)
	}
	defer tarReader.Close()

	// Create layer from tar archive
	layer, err := b.createLayer(tarReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create layer: %w", err)
	}

	// Create base image and add our layer
	img, err := b.buildImageWithLayer(layer, opts.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}

	// Save image to local storage
	storagePath, err := saveImageLocally(img, opts.Tag, b.storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to save image locally: %w", err)
	}

	// Update images map
	b.images[opts.Tag] = storagePath

	// Get image digest and size
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to get image digest: %w", err)
	}

	// Calculate size from storage file
	fileInfo, err := os.Stat(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}

	return &BuildResult{
		ImageID:     digest.String(),
		Digest:      digest,
		Size:        fileInfo.Size(),
		StoragePath: storagePath,
	}, nil
}

// createLayer creates an OCI layer from a tar reader
func (b *Builder) createLayer(tarReader io.Reader) (v1.Layer, error) {
	// Create layer from tar reader using go-containerregistry
	layer, err := tarball.LayerFromReader(tarReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create layer from tar: %w", err)
	}

	return layer, nil
}

// buildImageWithLayer creates an OCI image with a single layer and optional labels
func (b *Builder) buildImageWithLayer(layer v1.Layer, labels map[string]string) (v1.Image, error) {
	// Start with an empty image
	img := empty.Image

	// Add our layer to the image
	img, err := mutate.AppendLayers(img, layer)
	if err != nil {
		return nil, fmt.Errorf("failed to add layer to image: %w", err)
	}

	// If we have labels, add them to the config
	if len(labels) > 0 {
		configFile, err := img.ConfigFile()
		if err != nil {
			return nil, fmt.Errorf("failed to get config file: %w", err)
		}

		// Create a copy of the config and add labels
		newConfig := *configFile
		if newConfig.Config.Labels == nil {
			newConfig.Config.Labels = make(map[string]string)
		}

		for key, value := range labels {
			newConfig.Config.Labels[key] = value
		}

		// Add build timestamp
		newConfig.Config.Labels["org.opencontainers.image.created"] = time.Now().UTC().Format(time.RFC3339)

		// Update the image with new config
		img, err = mutate.ConfigFile(img, &newConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to update image config: %w", err)
		}
	}

	return img, nil
}

// GetStoragePath returns the local storage path for a tagged image
func (b *Builder) GetStoragePath(tag string) (string, bool) {
	path, exists := b.images[tag]
	return path, exists
}

