package registry

import (
	"bytes"
	"context"
	"io"
	
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Temporary stubs for missing dependencies until E2.1.1 (image-builder) is complete.
// These stubs allow the registry client to be tested and merged independently.

// ImageLoader interface defines image loading operations
// This will be implemented by E2.1.1 image-builder effort
type ImageLoader interface {
	LoadImage(ctx context.Context, source string) (v1.Image, error)
	ListImages(ctx context.Context) ([]string, error)
}

// MockImageLoader provides a stub implementation for testing
// This is a temporary implementation until the real ImageLoader from E2.1.1 is available
type MockImageLoader struct{}

// NewMockImageLoader creates a new mock image loader for testing
func NewMockImageLoader() ImageLoader {
	return &MockImageLoader{}
}

// LoadImage returns a test image for push operations
func (m *MockImageLoader) LoadImage(ctx context.Context, source string) (v1.Image, error) {
	// Create a minimal test image
	return &testImage{
		configFile: &v1.ConfigFile{
			Architecture: "amd64",
			OS:          "linux",
		},
	}, nil
}

// ListImages returns a test list of images
func (m *MockImageLoader) ListImages(ctx context.Context) ([]string, error) {
	return []string{
		"test/image:latest",
		"test/app:v1.0.0",
	}, nil
}

// testImage implements v1.Image for testing purposes
type testImage struct {
	configFile *v1.ConfigFile
}

// Layers returns empty layers for test image
func (t *testImage) Layers() ([]v1.Layer, error) {
	layer := &testLayer{}
	return []v1.Layer{layer}, nil
}

// MediaType returns the media type for test image
func (t *testImage) MediaType() (types.MediaType, error) {
	return types.DockerManifestSchema2, nil
}

// Size returns the size of test image
func (t *testImage) Size() (int64, error) {
	return 1024, nil
}

// ConfigName returns config hash for test image
func (t *testImage) ConfigName() (v1.Hash, error) {
	return v1.Hash{
		Algorithm: "sha256",
		Hex:       "test-config-hash",
	}, nil
}

// ConfigFile returns config file for test image
func (t *testImage) ConfigFile() (*v1.ConfigFile, error) {
	return t.configFile, nil
}

// RawConfigFile returns raw config for test image
func (t *testImage) RawConfigFile() ([]byte, error) {
	return []byte(`{"architecture":"amd64","os":"linux"}`), nil
}

// Digest returns digest for test image
func (t *testImage) Digest() (v1.Hash, error) {
	return v1.Hash{
		Algorithm: "sha256", 
		Hex:       "test-image-digest",
	}, nil
}

// Manifest returns manifest for test image
func (t *testImage) Manifest() (*v1.Manifest, error) {
	return &v1.Manifest{
		SchemaVersion: 2,
		MediaType:     types.DockerManifestSchema2,
		Config: v1.Descriptor{
			MediaType: types.DockerConfigJSON,
			Size:      100,
			Digest: v1.Hash{
				Algorithm: "sha256",
				Hex:       "test-config-hash",
			},
		},
		Layers: []v1.Descriptor{
			{
				MediaType: types.DockerLayer,
				Size:      1024,
				Digest: v1.Hash{
					Algorithm: "sha256",
					Hex:       "test-layer-hash",
				},
			},
		},
	}, nil
}

// RawManifest returns raw manifest for test image
func (t *testImage) RawManifest() ([]byte, error) {
	return []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json"}`), nil
}

// LayerByDigest returns a layer by digest for test image
func (t *testImage) LayerByDigest(h v1.Hash) (v1.Layer, error) {
	return &testLayer{}, nil
}

// LayerByDiffID returns a layer by diff ID for test image
func (t *testImage) LayerByDiffID(h v1.Hash) (v1.Layer, error) {
	return &testLayer{}, nil
}

// testLayer implements v1.Layer for testing
type testLayer struct{}

// Digest returns digest for test layer
func (l *testLayer) Digest() (v1.Hash, error) {
	return v1.Hash{
		Algorithm: "sha256",
		Hex:       "test-layer-digest",
	}, nil
}

// DiffID returns diff ID for test layer
func (l *testLayer) DiffID() (v1.Hash, error) {
	return v1.Hash{
		Algorithm: "sha256", 
		Hex:       "test-layer-diffid",
	}, nil
}

// Compressed returns compressed layer content
func (l *testLayer) Compressed() (io.ReadCloser, error) {
	data := []byte("test layer content")
	return io.NopCloser(bytes.NewReader(data)), nil
}

// Uncompressed returns uncompressed layer content  
func (l *testLayer) Uncompressed() (io.ReadCloser, error) {
	data := []byte("test layer content")
	return io.NopCloser(bytes.NewReader(data)), nil
}

// Size returns size of test layer
func (l *testLayer) Size() (int64, error) {
	return 1024, nil
}

// MediaType returns media type for test layer
func (l *testLayer) MediaType() (types.MediaType, error) {
	return types.DockerLayer, nil
}