package provider

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GatewayProviderStatus represents the duck-typed interface for Gateway providers
// All Gateway provider CRs must expose these fields in their status
type GatewayProviderStatus struct {
	// IngressClassName is the name of the ingress class to use in Ingress resources
	IngressClassName string

	// LoadBalancerEndpoint is the external endpoint for accessing services
	LoadBalancerEndpoint string

	// InternalEndpoint is the cluster-internal API endpoint
	InternalEndpoint string

	// Ready indicates whether the provider is ready
	Ready bool
}

// GetGatewayProviderStatus extracts duck-typed status from any Gateway provider CR
// It uses unstructured access to read the status fields that all Gateway providers must expose
func GetGatewayProviderStatus(obj *unstructured.Unstructured) (*GatewayProviderStatus, error) {
	if obj == nil {
		return nil, fmt.Errorf("object is nil")
	}

	status := &GatewayProviderStatus{}

	// Extract ingressClassName
	ingressClassName, found, err := unstructured.NestedString(obj.Object, "status", "ingressClassName")
	if err != nil {
		return nil, fmt.Errorf("failed to get ingressClassName: %w", err)
	}
	if found {
		status.IngressClassName = ingressClassName
	}

	// Extract loadBalancerEndpoint
	loadBalancerEndpoint, found, err := unstructured.NestedString(obj.Object, "status", "loadBalancerEndpoint")
	if err != nil {
		return nil, fmt.Errorf("failed to get loadBalancerEndpoint: %w", err)
	}
	if found {
		status.LoadBalancerEndpoint = loadBalancerEndpoint
	}

	// Extract internalEndpoint
	internalEndpoint, found, err := unstructured.NestedString(obj.Object, "status", "internalEndpoint")
	if err != nil {
		return nil, fmt.Errorf("failed to get internalEndpoint: %w", err)
	}
	if found {
		status.InternalEndpoint = internalEndpoint
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

// IsGatewayProviderReady checks if a Gateway provider is ready
func IsGatewayProviderReady(obj *unstructured.Unstructured) (bool, error) {
	status, err := GetGatewayProviderStatus(obj)
	if err != nil {
		return false, err
	}
	return status.Ready, nil
}
