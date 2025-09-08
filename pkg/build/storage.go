package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// saveImageLocally saves an OCI image to local storage as a tarball
func saveImageLocally(img v1.Image, tag, storageDir string) (string, error) {
	if img == nil {
		return "", fmt.Errorf("image cannot be nil")
	}

	if tag == "" {
		return "", fmt.Errorf("tag cannot be empty")
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Generate filename from tag (sanitize for filesystem)
	filename := sanitizeTagForFilename(tag) + ".tar"
	storagePath := filepath.Join(storageDir, filename)

	// Create the tarball file
	file, err := os.Create(storagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create tarball file: %w", err)
	}
	defer file.Close()

	// Write image as tarball using go-containerregistry
	if err := tarball.Write(nil, img, file); err != nil {
		// Clean up the file if write failed
		os.Remove(storagePath)
		return "", fmt.Errorf("failed to write image tarball: %w", err)
	}

	return storagePath, nil
}

// sanitizeTagForFilename converts a tag to a safe filename
func sanitizeTagForFilename(tag string) string {
	// Replace problematic characters with underscores
	safe := strings.ReplaceAll(tag, ":", "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, "<", "_")
	safe = strings.ReplaceAll(safe, ">", "_")
	safe = strings.ReplaceAll(safe, "\"", "_")
	safe = strings.ReplaceAll(safe, "|", "_")
	safe = strings.ReplaceAll(safe, "?", "_")
	safe = strings.ReplaceAll(safe, "*", "_")

	// Ensure it's not empty and not too long
	if safe == "" {
		safe = "unnamed"
	}
	
	// Truncate if too long
	if len(safe) > 100 {
		safe = safe[:100]
	}

	return safe
}