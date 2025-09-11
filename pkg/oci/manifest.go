package oci

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

// ParseManifest parses a manifest from JSON bytes and returns the appropriate type
func ParseManifest(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("manifest data is empty")
	}

	// First, determine the media type
	var rawManifest map[string]interface{}
	if err := json.Unmarshal(data, &rawManifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	mediaType, ok := rawManifest["mediaType"].(string)
	if !ok {
		return nil, fmt.Errorf("manifest missing mediaType field")
	}

	switch mediaType {
	case MediaTypeManifest, MediaTypeDockerManifest:
		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal image manifest: %w", err)
		}
		return &manifest, nil

	case MediaTypeManifestList, MediaTypeDockerManifestList:
		var manifestList ManifestList
		if err := json.Unmarshal(data, &manifestList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal manifest list: %w", err)
		}
		return &manifestList, nil

	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}
}

// ValidateManifest validates a manifest's structure and content
func ValidateManifest(manifest *Manifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema version: %d", manifest.SchemaVersion)
	}

	if manifest.MediaType == "" {
		return fmt.Errorf("manifest mediaType is required")
	}

	if !isValidMediaType(manifest.MediaType) {
		return fmt.Errorf("invalid manifest mediaType: %s", manifest.MediaType)
	}

	if err := validateDescriptor(&manifest.Config); err != nil {
		return fmt.Errorf("invalid config descriptor: %w", err)
	}

	if len(manifest.Layers) == 0 {
		return fmt.Errorf("manifest must have at least one layer")
	}

	for i, layer := range manifest.Layers {
		if err := validateDescriptor(&layer); err != nil {
			return fmt.Errorf("invalid layer %d descriptor: %w", i, err)
		}
	}

	return nil
}

// ValidateManifestList validates a manifest list's structure and content
func ValidateManifestList(manifestList *ManifestList) error {
	if manifestList == nil {
		return fmt.Errorf("manifest list is nil")
	}

	if manifestList.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema version: %d", manifestList.SchemaVersion)
	}

	if manifestList.MediaType == "" {
		return fmt.Errorf("manifest list mediaType is required")
	}

	if !isValidManifestListMediaType(manifestList.MediaType) {
		return fmt.Errorf("invalid manifest list mediaType: %s", manifestList.MediaType)
	}

	if len(manifestList.Manifests) == 0 {
		return fmt.Errorf("manifest list must have at least one manifest")
	}

	for i, manifest := range manifestList.Manifests {
		if err := validateDescriptor(&manifest); err != nil {
			return fmt.Errorf("invalid manifest %d descriptor: %w", i, err)
		}
	}

	return nil
}

// CreateManifest creates a new manifest with the given config and layers
func CreateManifest(config Descriptor, layers []Descriptor) *Manifest {
	return &Manifest{
		SchemaVersion: SchemaVersion,
		MediaType:     MediaTypeManifest,
		Config:        config,
		Layers:        layers,
		Annotations:   make(map[string]string),
	}
}

// CreateManifestList creates a new manifest list with the given manifests
func CreateManifestList(manifests []Descriptor) *ManifestList {
	return &ManifestList{
		SchemaVersion: SchemaVersion,
		MediaType:     MediaTypeManifestList,
		Manifests:     manifests,
		Annotations:   make(map[string]string),
	}
}

// ComputeDigest computes the SHA256 digest of the manifest
func ComputeDigest(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("cannot compute digest of empty data")
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash), nil
}

// GetManifestDigest computes the digest of a serialized manifest
func GetManifestDigest(manifest interface{}) (string, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return ComputeDigest(data)
}

// FilterManifestsByPlatform filters manifests by platform criteria
func FilterManifestsByPlatform(manifests []Descriptor, platform *Platform) []Descriptor {
	if platform == nil {
		return manifests
	}

	var filtered []Descriptor
	for _, manifest := range manifests {
		if manifest.Platform == nil {
			continue
		}

		if matchesPlatform(manifest.Platform, platform) {
			filtered = append(filtered, manifest)
		}
	}

	return filtered
}

// validateDescriptor validates a descriptor's structure
func validateDescriptor(desc *Descriptor) error {
	if desc.MediaType == "" {
		return fmt.Errorf("descriptor mediaType is required")
	}

	if desc.Digest == "" {
		return fmt.Errorf("descriptor digest is required")
	}

	if !strings.HasPrefix(desc.Digest, "sha256:") {
		return fmt.Errorf("descriptor digest must use sha256 algorithm")
	}

	if desc.Size <= 0 {
		return fmt.Errorf("descriptor size must be positive")
	}

	return nil
}

// isValidMediaType checks if the media type is valid for manifests
func isValidMediaType(mediaType string) bool {
	return mediaType == MediaTypeManifest || mediaType == MediaTypeDockerManifest
}

// isValidManifestListMediaType checks if the media type is valid for manifest lists
func isValidManifestListMediaType(mediaType string) bool {
	return mediaType == MediaTypeManifestList || mediaType == MediaTypeDockerManifestList
}

// matchesPlatform checks if two platforms match based on architecture and OS
func matchesPlatform(p1, p2 *Platform) bool {
	if p1 == nil || p2 == nil {
		return false
	}

	if p1.Architecture != p2.Architecture {
		return false
	}

	if p1.OS != p2.OS {
		return false
	}

	// Optional fields - if specified, they must match
	if p2.OSVersion != "" && p1.OSVersion != p2.OSVersion {
		return false
	}

	if p2.Variant != "" && p1.Variant != p2.Variant {
		return false
	}

	return true
}