package localbuild

import (
	"context"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		fieldName     string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid http URL",
			url:         "http://example.com",
			fieldName:   "test field",
			expectError: false,
		},
		{
			name:        "valid https URL",
			url:         "https://example.com:8443",
			fieldName:   "test field",
			expectError: false,
		},
		{
			name:          "empty URL",
			url:           "",
			fieldName:     "test field",
			expectError:   true,
			errorContains: "is not set",
		},
		{
			name:          "missing protocol",
			url:           "example.com",
			fieldName:     "test field",
			expectError:   true,
			errorContains: "must start with http:// or https://",
		},
		{
			name:          "http only",
			url:           "http://",
			fieldName:     "test field",
			expectError:   true,
			errorContains: "is too short",
		},
		{
			name:          "https only",
			url:           "https://",
			fieldName:     "test field",
			expectError:   true,
			errorContains: "is too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGitURL(tt.url, tt.fieldName)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReconcileGitRepoValidation(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)

	tests := []struct {
		name          string
		giteaProvider *v1alpha2.GiteaProvider
		expectError   bool
		errorContains string
	}{
		{
			name: "valid endpoints",
			giteaProvider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev-gitea",
					Namespace: util.GiteaNamespace,
				},
				Status: v1alpha2.GiteaProviderStatus{
					Endpoint:         "https://gitea.example.com:8443",
					InternalEndpoint: "http://gitea.svc.cluster.local:3000",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
						},
					},
					CredentialsSecretRef: &v1alpha2.SecretReference{
						Name:      "gitea-creds",
						Namespace: util.GiteaNamespace,
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty endpoint",
			giteaProvider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev-gitea",
					Namespace: util.GiteaNamespace,
				},
				Status: v1alpha2.GiteaProviderStatus{
					Endpoint:         "",
					InternalEndpoint: "http://gitea.svc.cluster.local:3000",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expectError:   true,
			errorContains: "endpoint is not set",
		},
		{
			name: "empty internal endpoint",
			giteaProvider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev-gitea",
					Namespace: util.GiteaNamespace,
				},
				Status: v1alpha2.GiteaProviderStatus{
					Endpoint:         "https://gitea.example.com:8443",
					InternalEndpoint: "",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expectError:   true,
			errorContains: "internal endpoint is not set",
		},
		{
			name: "invalid endpoint - missing protocol",
			giteaProvider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev-gitea",
					Namespace: util.GiteaNamespace,
				},
				Status: v1alpha2.GiteaProviderStatus{
					Endpoint:         "gitea.example.com:8443",
					InternalEndpoint: "http://gitea.svc.cluster.local:3000",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expectError:   true,
			errorContains: "endpoint must start with http:// or https://",
		},
		{
			name: "invalid endpoint - too short",
			giteaProvider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev-gitea",
					Namespace: util.GiteaNamespace,
				},
				Status: v1alpha2.GiteaProviderStatus{
					Endpoint:         "http://",
					InternalEndpoint: "http://gitea.svc.cluster.local:3000",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expectError:   true,
			errorContains: "endpoint is too short",
		},
		{
			name: "invalid endpoint - https too short",
			giteaProvider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev-gitea",
					Namespace: util.GiteaNamespace,
				},
				Status: v1alpha2.GiteaProviderStatus{
					Endpoint:         "https://",
					InternalEndpoint: "http://gitea.svc.cluster.local:3000",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expectError:   true,
			errorContains: "endpoint is too short",
		},
		{
			name: "provider not ready",
			giteaProvider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev-gitea",
					Namespace: util.GiteaNamespace,
				},
				Status: v1alpha2.GiteaProviderStatus{
					Endpoint:         "https://gitea.example.com:8443",
					InternalEndpoint: "http://gitea.svc.cluster.local:3000",
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			expectError:   true,
			errorContains: "is not ready yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.giteaProvider).
				WithStatusSubresource(tt.giteaProvider).
				Build()

			localbuild := &v1alpha1.Localbuild{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localdev",
					Namespace: globals.GetProjectNamespace("localdev"),
					Annotations: map[string]string{
						v1alpha1.CliStartTimeAnnotation: "2024-01-01T00:00:00Z",
					},
				},
				Spec: v1alpha1.LocalbuildSpec{
					BuildCustomization: v1alpha1.BuildCustomizationSpec{
						Protocol: "https",
						Host:     "example.com",
						Port:     "8443",
					},
				},
			}

			reconciler := &LocalbuildReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			ctx := context.Background()
			_, err := reconciler.reconcileGitRepo(ctx, localbuild, "embedded", "argocd", "argocd", "")

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
