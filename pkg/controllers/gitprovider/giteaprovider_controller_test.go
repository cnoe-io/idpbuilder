package gitprovider

import (
	"context"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGiteaProviderReconciler_Reconcile(t *testing.T) {
	t.Skip("Requires envtest environment for Gitea installation")
	scheme := k8s.GetScheme()

	tests := []struct {
		name            string
		provider        *v1alpha2.GiteaProvider
		expectedPhase   string
		expectFinalizer bool
		expectError     bool
		skipReconcile   bool
	}{
		{
			name: "new giteaprovider gets finalizer",
			provider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitea",
					Namespace: "gitea",
				},
				Spec: v1alpha2.GiteaProviderSpec{
					Namespace: "gitea",
					Version:   "1.24.3",
					AdminUser: v1alpha2.GiteaAdminUser{
						Username:     "giteaAdmin",
						Email:        "admin@test.com",
						AutoGenerate: true,
					},
				},
			},
			expectedPhase:   "", // We'll check after the first reconcile that adds finalizer
			expectFinalizer: true,
			expectError:     false,
			skipReconcile:   false, // Only do one reconcile to add finalizer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with objects
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(tt.provider).
				WithStatusSubresource(&v1alpha2.GiteaProvider{}).
				Build()

			reconciler := &GiteaProviderReconciler{
				Client: fakeClient,
				Scheme: scheme,
				Config: v1alpha1.BuildCustomizationSpec{
					Host:     "test.example.com",
					Port:     "8443",
					Protocol: "https",
				},
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.provider.Name,
					Namespace: tt.provider.Namespace,
				},
			}

			// First reconcile to add finalizer
			_, err := reconciler.Reconcile(context.Background(), req)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Get updated provider
			provider := &v1alpha2.GiteaProvider{}
			err = fakeClient.Get(context.Background(), req.NamespacedName, provider)
			require.NoError(t, err)

			// Verify finalizer
			if tt.expectFinalizer {
				assert.Contains(t, provider.Finalizers, giteaProviderFinalizer)
			}
		})
	}
}

func TestGiteaProviderReconciler_DeletionHandling(t *testing.T) {
	scheme := k8s.GetScheme()

	now := metav1.Now()
	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-gitea",
			Namespace:         "gitea",
			Finalizers:        []string{giteaProviderFinalizer},
			DeletionTimestamp: &now,
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(provider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
	}

	// Reconcile to handle deletion
	_, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)

	// Try to get updated provider - should not be found since it was deleted
	updatedProvider := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updatedProvider)
	// The object should be deleted (not found), or if still there, finalizer should be removed
	if err == nil {
		// If object is still there, finalizer should be removed
		assert.NotContains(t, updatedProvider.Finalizers, "giteaprovider.idpbuilder.cnoe.io/finalizer")
	}
}

func TestBuildConfigFromSpec(t *testing.T) {
	tests := []struct {
		name     string
		provider *v1alpha2.GiteaProvider
		expected v1alpha1.BuildCustomizationSpec
	}{
		{
			name: "all fields set in provider",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{
					Protocol:       "https",
					Host:           "example.com",
					Port:           "8443",
					UsePathRouting: true,
				},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "https",
				Host:           "example.com",
				IngressHost:    "example.com",
				Port:           "8443",
				UsePathRouting: true,
			},
		},
		{
			name: "empty fields use defaults",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "http",
				Host:           "cnoe.localtest.me",
				IngressHost:    "cnoe.localtest.me",
				Port:           "8080",
				UsePathRouting: false,
			},
		},
		{
			name: "partial fields with defaults",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{
					Host: "my-custom-host.com",
				},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "http",
				Host:           "my-custom-host.com",
				IngressHost:    "my-custom-host.com",
				Port:           "8080",
				UsePathRouting: false,
			},
		},
		{
			name: "IngressHost defaults to Host when empty",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{
					Protocol: "https",
					Host:     "test.localtest.me",
					Port:     "9443",
				},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "https",
				Host:           "test.localtest.me",
				IngressHost:    "test.localtest.me",
				Port:           "9443",
				UsePathRouting: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler := &GiteaProviderReconciler{}
			result := reconciler.buildConfigFromSpec(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}
