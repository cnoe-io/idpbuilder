package certs

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// parseCertificate parses certificate from PEM data
func parseCertificate(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("PEM block is not a certificate: %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// isFeatureEnabled checks if a feature flag is enabled
func isFeatureEnabled(flag string) bool {
	// Check environment variable
	envVar := fmt.Sprintf("IDPBUILDER_%s", flag)
	value := os.Getenv(envVar)

	// Parse boolean value
	return value == "true" || value == "1" || value == "enabled"
}

// findGiteaPod locates the Gitea pod in the cluster
func (e *KindCertExtractor) findGiteaPod(ctx context.Context, clusterName string) (string, error) {
	// Default namespace and labels for Gitea
	namespace := "gitea"
	labelSelector := "app=gitea"

	// Override from config if provided
	if e.config.Namespace != "" {
		namespace = e.config.Namespace
	}
	if e.config.PodLabelSelector != "" {
		labelSelector = e.config.PodLabelSelector
	}

	// Get pods matching selector
	pods, err := e.client.GetPods(ctx, namespace, labelSelector)
	if err != nil {
		return "", fmt.Errorf("failed to get pods: %w", err)
	}

	if len(pods) == 0 {
		return "", ErrGiteaPodNotFound
	}

	// Return first matching pod
	return pods[0], nil
}

// getClusterName retrieves the cluster name with fallback
func (e *KindCertExtractor) getClusterName() (string, error) {
	// Try configured name first
	if e.config.ClusterName != "" {
		return e.config.ClusterName, nil
	}

	// Fall back to current cluster
	return e.client.GetCurrentCluster()
}

// validateCertificateExpiry checks if certificate is valid time-wise
func validateCertificateExpiry(cert *x509.Certificate) error {
	now := time.Now()

	if now.Before(cert.NotBefore) {
		return ErrCertNotYetValid
	}

	if now.After(cert.NotAfter) {
		return ErrCertExpired
	}

	return nil
}

// expandHomeDir expands ~ to user home directory
func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// sanitizeRegistryName converts registry URL to safe filename
func sanitizeRegistryName(registry string) string {
	// Replace problematic characters
	safe := strings.ReplaceAll(registry, ":", "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, ".", "_")
	return safe
}