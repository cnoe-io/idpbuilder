package provider

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetProviderPhase extracts the phase from any provider CR
// All v1alpha2 provider CRs expose a phase field in their status
func GetProviderPhase(obj *unstructured.Unstructured) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("object is nil")
	}

	// Extract phase
	phase, found, err := unstructured.NestedString(obj.Object, "status", "phase")
	if err != nil {
		return "", fmt.Errorf("failed to get phase: %w", err)
	}
	if !found || phase == "" {
		return "Pending", nil
	}

	return phase, nil
}
