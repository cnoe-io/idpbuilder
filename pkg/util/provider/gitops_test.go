package provider

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetGitOpsProviderStatus(t *testing.T) {
	tests := []struct {
		name    string
		obj     *unstructured.Unstructured
		want    *GitOpsProviderStatus
		wantErr bool
	}{
		{
			name: "valid gitops provider",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "idpbuilder.cnoe.io/v1alpha2",
					"kind":       "ArgoCDProvider",
					"metadata": map[string]interface{}{
						"name":      "argocd",
						"namespace": "idpbuilder-system",
					},
					"status": map[string]interface{}{
						"endpoint":         "https://argocd.cnoe.localtest.me",
						"internalEndpoint": "http://argocd-server.argocd.svc.cluster.local",
						"credentialsSecretRef": map[string]interface{}{
							"name":      "argocd-admin-secret",
							"namespace": "argocd",
							"key":       "password",
						},
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			},
			want: &GitOpsProviderStatus{
				Endpoint:         "https://argocd.cnoe.localtest.me",
				InternalEndpoint: "http://argocd-server.argocd.svc.cluster.local",
				CredentialsSecretRef: SecretReference{
					Name:      "argocd-admin-secret",
					Namespace: "argocd",
					Key:       "password",
				},
				Ready: true,
			},
			wantErr: false,
		},
		{
			name: "gitops provider not ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"endpoint":         "https://argocd.cnoe.localtest.me",
						"internalEndpoint": "http://argocd-server.argocd.svc.cluster.local",
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			},
			want: &GitOpsProviderStatus{
				Endpoint:         "https://argocd.cnoe.localtest.me",
				InternalEndpoint: "http://argocd-server.argocd.svc.cluster.local",
				Ready:            false,
			},
			wantErr: false,
		},
		{
			name:    "nil object",
			obj:     nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGitOpsProviderStatus(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGitOpsProviderStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != nil && tt.want != nil {
				if got.Endpoint != tt.want.Endpoint {
					t.Errorf("Endpoint = %v, want %v", got.Endpoint, tt.want.Endpoint)
				}
				if got.InternalEndpoint != tt.want.InternalEndpoint {
					t.Errorf("InternalEndpoint = %v, want %v", got.InternalEndpoint, tt.want.InternalEndpoint)
				}
				if got.CredentialsSecretRef.Name != tt.want.CredentialsSecretRef.Name {
					t.Errorf("CredentialsSecretRef.Name = %v, want %v", got.CredentialsSecretRef.Name, tt.want.CredentialsSecretRef.Name)
				}
				if got.Ready != tt.want.Ready {
					t.Errorf("Ready = %v, want %v", got.Ready, tt.want.Ready)
				}
			}
		})
	}
}

func TestIsGitOpsProviderReady(t *testing.T) {
	tests := []struct {
		name    string
		obj     *unstructured.Unstructured
		want    bool
		wantErr bool
	}{
		{
			name: "ready gitops provider",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not ready gitops provider",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsGitOpsProviderReady(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsGitOpsProviderReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsGitOpsProviderReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
