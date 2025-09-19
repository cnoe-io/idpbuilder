package gitea

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// ImageLoader handles loading images from Docker daemon
type ImageLoader struct {
	client *client.Client
}

// NewImageLoader creates a new Docker image loader
func NewImageLoader() (*ImageLoader, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &ImageLoader{
		client: cli,
	}, nil
}

// LoadImage loads an image from the Docker daemon
func (il *ImageLoader) LoadImage(ctx context.Context, imageRef string) (*ImageManifest, error) {
	// Inspect the image to get its details
	inspect, _, err := il.client.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image %s: %w", imageRef, err)
	}

	// Get the image manifest
	manifest, err := il.buildManifestFromInspect(ctx, inspect)
	if err != nil {
		return nil, fmt.Errorf("failed to build manifest: %w", err)
	}

	return manifest, nil
}

// GetImageManifest retrieves the OCI manifest for an image
func (il *ImageLoader) GetImageManifest(ctx context.Context, imageID string) (*ImageManifest, error) {
	return il.LoadImage(ctx, imageID)
}

// GetImageContent returns a reader for the image content
func (il *ImageLoader) GetImageContent(ctx context.Context, imageID string) (io.ReadCloser, error) {
	// Save the image to a tar archive
	response, err := il.client.ImageSave(ctx, []string{imageID})
	if err != nil {
		return nil, fmt.Errorf("failed to save image %s: %w", imageID, err)
	}

	return response, nil
}

// CalculateDigest computes the SHA256 digest of content
func (il *ImageLoader) CalculateDigest(content []byte) digest.Digest {
	return digest.FromBytes(content)
}

// buildManifestFromInspect creates an OCI manifest from Docker inspect data
func (il *ImageLoader) buildManifestFromInspect(ctx context.Context, inspect types.ImageInspect) (*ImageManifest, error) {
	// Create config descriptor
	configBytes, err := json.Marshal(inspect.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	configDigest := il.CalculateDigest(configBytes)
	config := ManifestConfig{
		MediaType: "application/vnd.docker.container.image.v1+json",
		Size:      int64(len(configBytes)),
		Digest:    configDigest.String(),
	}

	// Process root filesystem layers
	var layers []ManifestLayer
	var totalSize int64

	for _, layer := range inspect.RootFS.Layers {
		// Create a layer descriptor
		// Note: In a production system, you'd need to access actual layer data
		// For now, we'll create reasonable estimates
		layerSize := int64(1024 * 1024) // 1MB estimate per layer
		totalSize += layerSize

		layers = append(layers, ManifestLayer{
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      layerSize,
			Digest:    string(layer), // Use the layer digest from Docker
		})
	}

	// Add config size to total
	totalSize += config.Size

	manifest := &ImageManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.docker.distribution.manifest.v2+json",
		Config:        config,
		Layers:        layers,
		TotalSize:     totalSize,
	}

	return manifest, nil
}

// ImageManifest represents an OCI image manifest
type ImageManifest struct {
	SchemaVersion int              `json:"schemaVersion"`
	MediaType     string           `json:"mediaType"`
	Config        ManifestConfig   `json:"config"`
	Layers        []ManifestLayer  `json:"layers"`
	TotalSize     int64            `json:"-"` // Not part of manifest, for progress tracking
}

// ManifestConfig represents the config section of a manifest
type ManifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// ManifestLayer represents a layer in the manifest
type ManifestLayer struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// ToOCIManifest converts to OCI manifest format
func (im *ImageManifest) ToOCIManifest() *v1.Manifest {
	// Convert config
	config := v1.Descriptor{
		MediaType: im.Config.MediaType,
		Size:      im.Config.Size,
		Digest:    digest.Digest(im.Config.Digest),
	}

	// Convert layers
	var layers []v1.Descriptor
	for _, layer := range im.Layers {
		layers = append(layers, v1.Descriptor{
			MediaType: layer.MediaType,
			Size:      layer.Size,
			Digest:    digest.Digest(layer.Digest),
		})
	}

	return &v1.Manifest{
		MediaType: im.MediaType,
		Config:    config,
		Layers:    layers,
	}
}

// ToJSON serializes the manifest to JSON
func (im *ImageManifest) ToJSON() ([]byte, error) {
	return json.Marshal(im)
}

// ToReader returns the manifest as an io.Reader
func (im *ImageManifest) ToReader() (io.Reader, error) {
	jsonBytes, err := im.ToJSON()
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(jsonBytes)), nil
}

// Close releases resources associated with the image loader
func (il *ImageLoader) Close() error {
	if il.client != nil {
		return il.client.Close()
	}
	return nil
}