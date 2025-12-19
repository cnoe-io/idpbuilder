package provider

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetGitProviderStatus(t *testing.T) {
	tests := []struct {
		name    string
		obj     *unstructured.Unstructured
		want    *GitProviderStatus
		wantErr bool
	}{
		{
			name:    "nil object",
			obj:     nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "complete status",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"endpoint":         "https://gitea.example.com",
						"internalEndpoint": "http://gitea.svc.cluster.local:3000",
						"credentialsSecretRef": map[string]interface{}{
							"name":      "gitea-creds",
							"namespace": "gitea",
							"key":       "token",
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
			want: &GitProviderStatus{
				Endpoint:         "https://gitea.example.com",
				InternalEndpoint: "http://gitea.svc.cluster.local:3000",
				CredentialsSecretRef: SecretReference{
					Name:      "gitea-creds",
					Namespace: "gitea",
					Key:       "token",
				},
				Ready: true,
			},
			wantErr: false,
		},
		{
			name: "partial status - no credentials",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"endpoint":         "https://gitea.example.com",
						"internalEndpoint": "http://gitea.svc.cluster.local:3000",
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			},
			want: &GitProviderStatus{
				Endpoint:         "https://gitea.example.com",
				InternalEndpoint: "http://gitea.svc.cluster.local:3000",
				Ready:            false,
			},
			wantErr: false,
		},
		{
			name: "not ready - missing Ready condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"endpoint":         "https://gitea.example.com",
						"internalEndpoint": "http://gitea.svc.cluster.local:3000",
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Installing",
								"status": "True",
							},
						},
					},
				},
			},
			want: &GitProviderStatus{
				Endpoint:         "https://gitea.example.com",
				InternalEndpoint: "http://gitea.svc.cluster.local:3000",
				Ready:            false,
			},
			wantErr: false,
		},
		{
			name: "empty status",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			want: &GitProviderStatus{
				Ready: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGitProviderStatus(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGitProviderStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Endpoint != tt.want.Endpoint {
				t.Errorf("GetGitProviderStatus().Endpoint = %v, want %v", got.Endpoint, tt.want.Endpoint)
			}
			if got.InternalEndpoint != tt.want.InternalEndpoint {
				t.Errorf("GetGitProviderStatus().InternalEndpoint = %v, want %v", got.InternalEndpoint, tt.want.InternalEndpoint)
			}
			if got.Ready != tt.want.Ready {
				t.Errorf("GetGitProviderStatus().Ready = %v, want %v", got.Ready, tt.want.Ready)
			}
			if got.CredentialsSecretRef.Name != tt.want.CredentialsSecretRef.Name {
				t.Errorf("GetGitProviderStatus().CredentialsSecretRef.Name = %v, want %v",
					got.CredentialsSecretRef.Name, tt.want.CredentialsSecretRef.Name)
			}
		})
	}
}

func TestIsGitProviderReady(t *testing.T) {
	tests := []struct {
		name    string
		obj     *unstructured.Unstructured
		want    bool
		wantErr bool
	}{
		{
			name:    "nil object",
			obj:     nil,
			want:    false,
			wantErr: true,
		},
		{
			name: "ready provider",
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
			name: "not ready provider",
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
			got, err := IsGitProviderReady(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsGitProviderReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsGitProviderReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
