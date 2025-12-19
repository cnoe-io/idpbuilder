package platform

import (
	"context"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPlatformReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha2.AddToScheme(scheme))

	tests := []struct {
		name          string
		platform      *v1alpha2.Platform
		providers     []runtime.Object
		expectedPhase string
		expectedReady bool
		expectRequeue bool
	}{
		{
			name: "platform with ready gitea provider",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "default",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.example.com",
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "my-gitea",
								Kind:      "GiteaProvider",
								Namespace: "gitea",
							},
						},
					},
				},
			},
			providers: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "idpbuilder.cnoe.io/v1alpha2",
						"kind":       "GiteaProvider",
						"metadata": map[string]interface{}{
							"name":      "my-gitea",
							"namespace": "gitea",
						},
						"status": map[string]interface{}{
							"endpoint":         "https://gitea.example.com",
							"internalEndpoint": "http://gitea.svc:3000",
							"conditions": []interface{}{
								map[string]interface{}{
									"type":   "Ready",
									"status": "True",
								},
							},
						},
					},
				},
			},
			expectedPhase: "Ready",
			expectedReady: true,
			expectRequeue: false,
		},
		{
			name: "platform with not ready gitea provider",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "default",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.example.com",
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "my-gitea",
								Kind:      "GiteaProvider",
								Namespace: "gitea",
							},
						},
					},
				},
			},
			providers: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "idpbuilder.cnoe.io/v1alpha2",
						"kind":       "GiteaProvider",
						"metadata": map[string]interface{}{
							"name":      "my-gitea",
							"namespace": "gitea",
						},
						"status": map[string]interface{}{
							"endpoint":         "https://gitea.example.com",
							"internalEndpoint": "http://gitea.svc:3000",
							"conditions": []interface{}{
								map[string]interface{}{
									"type":   "Ready",
									"status": "False",
								},
							},
						},
					},
				},
			},
			expectedPhase: "Pending",
			expectedReady: false,
			expectRequeue: true,
		},
		{
			name: "platform with missing provider",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "default",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.example.com",
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "my-gitea",
								Kind:      "GiteaProvider",
								Namespace: "gitea",
							},
						},
					},
				},
			},
			providers:     []runtime.Object{},
			expectedPhase: "Pending",
			expectedReady: false,
			expectRequeue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with objects
			objs := append([]runtime.Object{tt.platform}, tt.providers...)
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objs...).
				WithStatusSubresource(&v1alpha2.Platform{}).
				Build()

			reconciler := &PlatformReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.platform.Name,
					Namespace: tt.platform.Namespace,
				},
			}

			// First reconcile to add finalizer
			_, err := reconciler.Reconcile(context.Background(), req)
			require.NoError(t, err)

			// Second reconcile to process
			result, err := reconciler.Reconcile(context.Background(), req)
			require.NoError(t, err)

			// Verify requeue expectation
			if tt.expectRequeue {
				assert.True(t, result.RequeueAfter > 0, "Expected requeue after delay")
			} else {
				assert.Equal(t, ctrl.Result{}, result, "Expected no requeue")
			}

			// Get updated platform
			platform := &v1alpha2.Platform{}
			err = fakeClient.Get(context.Background(), req.NamespacedName, platform)
			require.NoError(t, err)

			// Verify phase
			assert.Equal(t, tt.expectedPhase, platform.Status.Phase)

			// Verify ready condition
			readyCondition := findCondition(platform.Status.Conditions, "Ready")
			if tt.expectedReady {
				require.NotNil(t, readyCondition, "Ready condition should exist")
				assert.Equal(t, metav1.ConditionTrue, readyCondition.Status)
			} else {
				if readyCondition != nil {
					assert.Equal(t, metav1.ConditionFalse, readyCondition.Status)
				}
			}
		})
	}
}

func TestPlatformReconciler_ReconcileGitProviders(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha2.AddToScheme(scheme))

	readyProvider := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "idpbuilder.cnoe.io/v1alpha2",
			"kind":       "GiteaProvider",
			"metadata": map[string]interface{}{
				"name":      "provider1",
				"namespace": "gitea",
			},
			"status": map[string]interface{}{
				"endpoint":         "https://gitea1.example.com",
				"internalEndpoint": "http://gitea1.svc:3000",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	notReadyProvider := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "idpbuilder.cnoe.io/v1alpha2",
			"kind":       "GiteaProvider",
			"metadata": map[string]interface{}{
				"name":      "provider2",
				"namespace": "gitea",
			},
			"status": map[string]interface{}{
				"endpoint":         "https://gitea2.example.com",
				"internalEndpoint": "http://gitea2.svc:3000",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "False",
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(readyProvider, notReadyProvider).
		Build()

	reconciler := &PlatformReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	platform := &v1alpha2.Platform{
		Spec: v1alpha2.PlatformSpec{
			Components: v1alpha2.PlatformComponents{
				GitProviders: []v1alpha2.ProviderReference{
					{Name: "provider1", Kind: "GiteaProvider", Namespace: "gitea"},
					{Name: "provider2", Kind: "GiteaProvider", Namespace: "gitea"},
				},
			},
		},
	}

	summary, allReady, err := reconciler.reconcileGitProviders(context.Background(), platform)
	require.NoError(t, err)
	assert.Len(t, summary, 2)
	assert.False(t, allReady, "Not all providers should be ready")

	// Check individual summaries
	provider1Summary := findProviderSummary(summary, "provider1")
	require.NotNil(t, provider1Summary)
	assert.True(t, provider1Summary.Ready)

	provider2Summary := findProviderSummary(summary, "provider2")
	require.NotNil(t, provider2Summary)
	assert.False(t, provider2Summary.Ready)
}

// Helper functions
func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

func findProviderSummary(summaries []v1alpha2.ProviderStatusSummary, name string) *v1alpha2.ProviderStatusSummary {
	for i := range summaries {
		if summaries[i].Name == name {
			return &summaries[i]
		}
	}
	return nil
}
