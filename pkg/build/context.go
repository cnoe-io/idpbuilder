package build

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// createTarFromContext creates a tar archive from a directory context with exclusions
func createTarFromContext(contextPath string, exclusions []string) (io.ReadCloser, error) {
	// Validate context directory exists
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("context directory does not exist: %s", contextPath)
	}

	// Create pipe for streaming tar data
	pr, pw := io.Pipe()

	// Start goroutine to write tar data
	go func() {
		defer pw.Close()

		tw := tar.NewWriter(pw)
		defer tw.Close()

		// Walk directory tree and add files to tar
		err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path from context root
			relPath, err := filepath.Rel(contextPath, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			// Skip the root directory itself
			if relPath == "." {
				return nil
			}

			// Check if file should be excluded
			if shouldExclude(relPath, exclusions) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return fmt.Errorf("failed to create tar header for %s: %w", path, err)
			}

			// Use relative path in tar
			header.Name = relPath

			// Write header
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write tar header for %s: %w", path, err)
			}

			// Write file content if it's a regular file
			if info.Mode().IsRegular() {
				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open file %s: %w", path, err)
				}
				defer file.Close()

				if _, err := io.Copy(tw, file); err != nil {
					return fmt.Errorf("failed to copy file content for %s: %w", path, err)
				}
			}

			return nil
		})

		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create tar from context: %w", err))
		}
	}()

	return pr, nil
}

// shouldExclude checks if a path matches any of the exclusion patterns
func shouldExclude(path string, exclusions []string) bool {
	for _, pattern := range exclusions {
		// Simple pattern matching - supports basic glob patterns
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Also check if pattern matches any part of the path
		pathParts := strings.Split(path, string(filepath.Separator))
		for _, part := range pathParts {
			if matched, err := filepath.Match(pattern, part); err == nil && matched {
				return true
			}
		}

		// Check if the path starts with the pattern (for directory exclusions)
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}