package provider

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GitProviderStatus represents the duck-typed interface for Git providers
// All Git provider CRs must expose these fields in their status
type GitProviderStatus struct {
	// Endpoint is the external URL for Git web UI and cloning
	Endpoint string

	// InternalEndpoint is the cluster-internal URL for API access
	InternalEndpoint string

	// CredentialsSecretRef references the secret containing access credentials
	CredentialsSecretRef SecretReference

	// Ready indicates whether the provider is ready
	Ready bool
}

// SecretReference contains information to locate a secret
type SecretReference struct {
	Name      string
	Namespace string
	Key       string
}

// GetGitProviderStatus extracts duck-typed status from any Git provider CR
// It uses unstructured access to read the status fields that all Git providers must expose
func GetGitProviderStatus(obj *unstructured.Unstructured) (*GitProviderStatus, error) {
	if obj == nil {
		return nil, fmt.Errorf("object is nil")
	}

	status := &GitProviderStatus{}

	// Extract endpoint
	endpoint, found, err := unstructured.NestedString(obj.Object, "status", "endpoint")
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}
	if found {
		status.Endpoint = endpoint
	}

	// Extract internalEndpoint
	internalEndpoint, found, err := unstructured.NestedString(obj.Object, "status", "internalEndpoint")
	if err != nil {
		return nil, fmt.Errorf("failed to get internalEndpoint: %w", err)
	}
	if found {
		status.InternalEndpoint = internalEndpoint
	}

	// Extract credentialsSecretRef
	credRef, found, err := unstructured.NestedMap(obj.Object, "status", "credentialsSecretRef")
	if err != nil {
		return nil, fmt.Errorf("failed to get credentialsSecretRef: %w", err)
	}
	if found {
		if name, ok := credRef["name"].(string); ok {
			status.CredentialsSecretRef.Name = name
		}
		if namespace, ok := credRef["namespace"].(string); ok {
			status.CredentialsSecretRef.Namespace = namespace
		}
		if key, ok := credRef["key"].(string); ok {
			status.CredentialsSecretRef.Key = key
		}
	}

	// Determine if provider is ready by checking the Ready condition
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil {
		return nil, fmt.Errorf("failed to get conditions: %w", err)
	}
	if found {
		for _, condition := range conditions {
			condMap, ok := condition.(map[string]interface{})
			if !ok {
				continue
			}
			condType, ok := condMap["type"].(string)
			if !ok || condType != "Ready" {
				continue
			}
			condStatus, ok := condMap["status"].(string)
			if ok && condStatus == "True" {
				status.Ready = true
				break
			}
		}
	}

	return status, nil
}

// IsGitProviderReady checks if a Git provider is ready
func IsGitProviderReady(obj *unstructured.Unstructured) (bool, error) {
	status, err := GetGitProviderStatus(obj)
	if err != nil {
		return false, err
	}
	return status.Ready, nil
}
