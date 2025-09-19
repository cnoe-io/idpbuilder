package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

var (
	// imageRefPattern matches standard image references
	// Format: [registry[:port]/][namespace/]repository[:tag][@digest]
	imageRefPattern = regexp.MustCompile(`^(?:([^/]+(?:\.[^/]*)*(?:\:\d+)?)/)?(?:([^/]+)/)?([^/:@]+)(?::([^@]+))?(?:@(.+))?$`)

	// tagPattern validates tags
	tagPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// digestPattern validates digest format
	digestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

// ParseImageReference parses an image reference string into components
func ParseImageReference(ref string) (*types.ImageReference, error) {
	if ref == "" {
		return nil, fmt.Errorf("image reference cannot be empty")
	}

	matches := imageRefPattern.FindStringSubmatch(ref)
	if len(matches) != 6 {
		return nil, fmt.Errorf("invalid image reference format: %s", ref)
	}

	registry := matches[1]
	namespace := matches[2]
	repository := matches[3]
	tag := matches[4]
	digest := matches[5]

	// Default registry
	if registry == "" {
		registry = "docker.io"
	}

	// Default namespace for Docker Hub
	if registry == "docker.io" && namespace == "" {
		namespace = "library"
	}

	// Handle case where first part might be namespace, not registry
	// (e.g., "nginx/nginx" should be "docker.io/nginx/nginx")
	if registry != "" && namespace == "" && !strings.Contains(registry, ".") && !strings.Contains(registry, ":") {
		namespace = registry
		registry = "docker.io"
	}

	// Default tag
	if tag == "" && digest == "" {
		tag = "latest"
	}

	// Validate tag format
	if tag != "" && !tagPattern.MatchString(tag) {
		return nil, fmt.Errorf("invalid tag format: %s", tag)
	}

	// Validate digest format
	if digest != "" && !digestPattern.MatchString(digest) {
		return nil, fmt.Errorf("invalid digest format: %s", digest)
	}

	return &types.ImageReference{
		Registry:   registry,
		Namespace:  namespace,
		Repository: repository,
		Tag:        tag,
		Digest:     digest,
	}, nil
}

// BuildImageReference constructs an image reference string from components
func BuildImageReference(ref *types.ImageReference) string {
	if ref == nil || ref.Repository == "" {
		return ""
	}

	var parts []string

	// Add registry if not Docker Hub
	if ref.Registry != "" && ref.Registry != "docker.io" {
		parts = append(parts, ref.Registry)
	}

	// Add namespace/repository
	if ref.Namespace != "" && !(ref.Registry == "docker.io" && ref.Namespace == "library") {
		parts = append(parts, ref.Namespace, ref.Repository)
	} else {
		parts = append(parts, ref.Repository)
	}

	image := strings.Join(parts, "/")

	// Add tag
	if ref.Tag != "" {
		image = fmt.Sprintf("%s:%s", image, ref.Tag)
	}

	// Add digest (overrides tag)
	if ref.Digest != "" {
		if ref.Tag != "" {
			image = fmt.Sprintf("%s@%s", image, ref.Digest)
		} else {
			// Remove any existing tag when using digest
			if idx := strings.LastIndex(image, ":"); idx > 0 && !strings.Contains(image[idx:], "/") {
				image = image[:idx]
			}
			image = fmt.Sprintf("%s@%s", image, ref.Digest)
		}
	}

	return image
}

// ValidateImageReference validates an image reference
func ValidateImageReference(ref *types.ImageReference) error {
	if ref == nil {
		return fmt.Errorf("image reference cannot be nil")
	}

	if ref.Repository == "" {
		return fmt.Errorf("repository name is required")
	}

	// Validate repository name format
	if !regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?$`).MatchString(ref.Repository) {
		return fmt.Errorf("invalid repository name format: %s", ref.Repository)
	}

	// Validate tag if present
	if ref.Tag != "" && !tagPattern.MatchString(ref.Tag) {
		return fmt.Errorf("invalid tag format: %s", ref.Tag)
	}

	// Validate digest if present
	if ref.Digest != "" && !digestPattern.MatchString(ref.Digest) {
		return fmt.Errorf("invalid digest format: %s", ref.Digest)
	}

	return nil
}

// ImageReferenceWithTag creates a new ImageReference with the specified tag
func ImageReferenceWithTag(ref *types.ImageReference, tag string) *types.ImageReference {
	if ref == nil {
		return nil
	}

	newRef := *ref // Copy
	newRef.Tag = tag
	newRef.Digest = "" // Clear digest when setting tag
	return &newRef
}

// ImageReferenceWithDigest creates a new ImageReference with the specified digest
func ImageReferenceWithDigest(ref *types.ImageReference, digest string) *types.ImageReference {
	if ref == nil {
		return nil
	}

	newRef := *ref // Copy
	newRef.Digest = digest
	// Keep tag as it can coexist with digest
	return &newRef
}

// GetImageRegistryURL extracts the registry URL from an image reference
func GetImageRegistryURL(ref *types.ImageReference) string {
	if ref == nil || ref.Registry == "" {
		return "https://registry-1.docker.io" // Default Docker Hub
	}

	// Handle special case for Docker Hub
	if ref.Registry == "docker.io" {
		return "https://registry-1.docker.io"
	}

	// Default to HTTPS if no scheme specified
	if !strings.HasPrefix(ref.Registry, "http://") && !strings.HasPrefix(ref.Registry, "https://") {
		return "https://" + ref.Registry
	}

	return ref.Registry
}