package provider

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetProviderPhase(t *testing.T) {
	tests := []struct {
		name          string
		obj           *unstructured.Unstructured
		expectedPhase string
		expectError   bool
	}{
		{
			name:        "nil object",
			obj:         nil,
			expectError: true,
		},
		{
			name: "object with Installing phase",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"phase": "Installing",
					},
				},
			},
			expectedPhase: "Installing",
			expectError:   false,
		},
		{
			name: "object with Ready phase",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"phase": "Ready",
					},
				},
			},
			expectedPhase: "Ready",
			expectError:   false,
		},
		{
			name: "object with Failed phase",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"phase": "Failed",
					},
				},
			},
			expectedPhase: "Failed",
			expectError:   false,
		},
		{
			name: "object with no phase defaults to Pending",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{},
				},
			},
			expectedPhase: "Pending",
			expectError:   false,
		},
		{
			name: "object with empty phase defaults to Pending",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"phase": "",
					},
				},
			},
			expectedPhase: "Pending",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase, err := GetProviderPhase(tt.obj)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if phase != tt.expectedPhase {
				t.Errorf("Expected phase %s, got %s", tt.expectedPhase, phase)
			}
		})
	}
}
