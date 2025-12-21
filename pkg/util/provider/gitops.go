package provider

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GitOpsProviderStatus represents the duck-typed interface for GitOps providers
// All GitOps provider CRs must expose these fields in their status
type GitOpsProviderStatus struct {
	// Endpoint is the external URL for GitOps provider web UI
	Endpoint string

	// InternalEndpoint is the cluster-internal URL for API access
	InternalEndpoint string

	// CredentialsSecretRef references the secret containing access credentials
	CredentialsSecretRef SecretReference

	// Ready indicates whether the provider is ready
	Ready bool
}

// GetGitOpsProviderStatus extracts duck-typed status from any GitOps provider CR
// It uses unstructured access to read the status fields that all GitOps providers must expose
func GetGitOpsProviderStatus(obj *unstructured.Unstructured) (*GitOpsProviderStatus, error) {
	if obj == nil {
		return nil, fmt.Errorf("object is nil")
	}

	status := &GitOpsProviderStatus{}

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

// IsGitOpsProviderReady checks if a GitOps provider is ready
func IsGitOpsProviderReady(obj *unstructured.Unstructured) (bool, error) {
	status, err := GetGitOpsProviderStatus(obj)
	if err != nil {
		return false, err
	}
	return status.Ready, nil
}
