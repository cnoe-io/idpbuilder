//go:build integration
// +build integration

package gitprovider_test

import (
	"context"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/gitprovider"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/platform"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestPlatformGiteaProviderWorkflow validates the complete workflow:
// 1. Platform CR is created
// 2. Platform references a GiteaProvider CR
// 3. GiteaProvider controller creates Gitea deployment
// 4. Platform controller aggregates GiteaProvider status
// NOTE: This test is skipped because it requires envtest or a real cluster
// to properly test the Gitea deployment. The unit tests below cover the
// critical workflow validation.
func TestPlatformGiteaProviderWorkflow(t *testing.T) {
	t.Skip("Requires envtest environment for full Gitea installation")
	scheme := k8s.GetScheme()

	ctx := context.Background()

	// Step 1: Create GiteaProvider CR
	giteaProvider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
			AdminUser: v1alpha2.GiteaAdminUser{
				Username:     "giteaAdmin",
				Email:        "admin@cnoe.localtest.me",
				AutoGenerate: true,
			},
		},
	}

	// Step 2: Create Platform CR that references the GiteaProvider
	platformCR := &v1alpha2.Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-platform",
			Namespace: "default",
		},
		Spec: v1alpha2.PlatformSpec{
			Domain: "cnoe.localtest.me",
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
	}

	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gitea",
		},
	}

	// Create fake client with objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(namespace, giteaProvider, platformCR).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}, &v1alpha2.Platform{}).
		Build()

	config := v1alpha1.BuildCustomizationSpec{
		Host:     "cnoe.localtest.me",
		Port:     "8443",
		Protocol: "https",
	}

	// Create GiteaProvider reconciler
	giteaReconciler := &gitprovider.GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: config,
	}

	// Create Platform reconciler
	platformReconciler := &platform.PlatformReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Step 3: Reconcile GiteaProvider
	giteaReq := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      giteaProvider.Name,
			Namespace: giteaProvider.Namespace,
		},
	}

	// First reconcile to add finalizer and set initial status
	_, err := giteaReconciler.Reconcile(ctx, giteaReq)
	require.NoError(t, err)

	// Verify GiteaProvider has finalizer and initial phase
	updatedGitea := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(ctx, giteaReq.NamespacedName, updatedGitea)
	require.NoError(t, err)
	assert.Contains(t, updatedGitea.Finalizers, "giteaprovider.idpbuilder.cnoe.io/finalizer")
	assert.Equal(t, "Installing", updatedGitea.Status.Phase)

	// Step 4: Reconcile Platform
	platformReq := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      platformCR.Name,
			Namespace: platformCR.Namespace,
		},
	}

	// First reconcile to add finalizer
	_, err = platformReconciler.Reconcile(ctx, platformReq)
	require.NoError(t, err)

	// Verify Platform has finalizer
	updatedPlatform := &v1alpha2.Platform{}
	err = fakeClient.Get(ctx, platformReq.NamespacedName, updatedPlatform)
	require.NoError(t, err)
	assert.Contains(t, updatedPlatform.Finalizers, "platform.idpbuilder.cnoe.io/finalizer")

	// Second reconcile to process providers
	result, err := platformReconciler.Reconcile(ctx, platformReq)
	require.NoError(t, err)

	// Since GiteaProvider is not ready yet, Platform should requeue
	assert.True(t, result.RequeueAfter > 0, "Platform should requeue when GiteaProvider is not ready")

	// Get updated platform status
	err = fakeClient.Get(ctx, platformReq.NamespacedName, updatedPlatform)
	require.NoError(t, err)

	// Verify Platform references the GiteaProvider
	assert.Len(t, updatedPlatform.Status.Providers.GitProviders, 1)
	assert.Equal(t, "my-gitea", updatedPlatform.Status.Providers.GitProviders[0].Name)
	assert.Equal(t, "GiteaProvider", updatedPlatform.Status.Providers.GitProviders[0].Kind)
	assert.False(t, updatedPlatform.Status.Providers.GitProviders[0].Ready, "GiteaProvider should not be ready yet")

	// Verify Platform status
	assert.Equal(t, "Pending", updatedPlatform.Status.Phase)
	assert.Len(t, updatedPlatform.Status.Conditions, 1)
	assert.Equal(t, "Ready", updatedPlatform.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionFalse, updatedPlatform.Status.Conditions[0].Status)
}

// TestPlatformWithoutGiteaProvider ensures Platform doesn't directly create Gitea
func TestPlatformWithoutGiteaProvider(t *testing.T) {
	scheme := k8s.GetScheme()

	ctx := context.Background()

	// Create Platform CR without any GiteaProvider reference
	platformCR := &v1alpha2.Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "platform-no-gitea",
			Namespace: "default",
		},
		Spec: v1alpha2.PlatformSpec{
			Domain: "cnoe.localtest.me",
			Components: v1alpha2.PlatformComponents{
				GitProviders: []v1alpha2.ProviderReference{},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(platformCR).
		WithStatusSubresource(&v1alpha2.Platform{}).
		Build()

	platformReconciler := &platform.PlatformReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      platformCR.Name,
			Namespace: platformCR.Namespace,
		},
	}

	// Reconcile Platform twice (finalizer + process)
	_, err := platformReconciler.Reconcile(ctx, req)
	require.NoError(t, err)

	_, err = platformReconciler.Reconcile(ctx, req)
	require.NoError(t, err)

	// Verify Platform is initializing (no providers configured)
	updatedPlatform := &v1alpha2.Platform{}
	err = fakeClient.Get(ctx, req.NamespacedName, updatedPlatform)
	require.NoError(t, err)

	// In v2, Platform without providers is "Initializing" not "Ready"
	assert.Equal(t, "Initializing", updatedPlatform.Status.Phase)
	assert.Len(t, updatedPlatform.Status.Providers.GitProviders, 0)
}

// TestGiteaProviderCreatesDeployment validates that GiteaProvider controller
// is responsible for creating the Gitea deployment, not the old LocalbuildReconciler
// NOTE: This test is skipped because it requires envtest or a real cluster
// to properly test the Gitea deployment.
func TestGiteaProviderCreatesDeployment(t *testing.T) {
	t.Skip("Requires envtest environment for Gitea installation")
	scheme := k8s.GetScheme()

	ctx := context.Background()

	giteaProvider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gitea",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(namespace, giteaProvider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &gitprovider.GiteaProviderReconciler{
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
			Name:      giteaProvider.Name,
			Namespace: giteaProvider.Namespace,
		},
	}

	// Reconcile to initialize and install Gitea
	_, err := reconciler.Reconcile(ctx, req)
	require.NoError(t, err)

	// Verify GiteaProvider status was updated
	updated := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(ctx, req.NamespacedName, updated)
	require.NoError(t, err)

	// The GiteaProvider should be in Installing phase
	assert.Equal(t, "Installing", updated.Status.Phase)

	// Verify that the reconciler attempted to install resources
	// (In a real environment, this would create deployment, service, etc.)
	// For unit test, we just verify the controller set the correct phase
}

// TestPlatformStatusAggregation validates that Platform aggregates status from GiteaProvider
func TestPlatformStatusAggregation(t *testing.T) {
	scheme := k8s.GetScheme()

	ctx := context.Background()

	// Create a ready GiteaProvider
	giteaProvider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ready-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
		Status: v1alpha2.GiteaProviderStatus{
			Phase:            "Ready",
			Endpoint:         "https://gitea.test.com",
			InternalEndpoint: "http://my-gitea-http.gitea.svc.cluster.local:3000",
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             "GiteaReady",
					Message:            "Gitea is ready",
				},
			},
		},
	}

	platformCR := &v1alpha2.Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-platform",
			Namespace: "default",
		},
		Spec: v1alpha2.PlatformSpec{
			Domain: "test.com",
			Components: v1alpha2.PlatformComponents{
				GitProviders: []v1alpha2.ProviderReference{
					{
						Name:      "ready-gitea",
						Kind:      "GiteaProvider",
						Namespace: "gitea",
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(giteaProvider, platformCR).
		WithStatusSubresource(&v1alpha2.Platform{}, &v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &platform.PlatformReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      platformCR.Name,
			Namespace: platformCR.Namespace,
		},
	}

	// Reconcile to add finalizer
	_, err := reconciler.Reconcile(ctx, req)
	require.NoError(t, err)

	// Reconcile to process providers
	result, err := reconciler.Reconcile(ctx, req)
	require.NoError(t, err)

	// Should not requeue since provider is ready
	assert.Equal(t, ctrl.Result{}, result)

	// Verify Platform status
	updated := &v1alpha2.Platform{}
	err = fakeClient.Get(ctx, req.NamespacedName, updated)
	require.NoError(t, err)

	assert.Equal(t, "Ready", updated.Status.Phase)
	assert.Len(t, updated.Status.Providers.GitProviders, 1)
	assert.Equal(t, "ready-gitea", updated.Status.Providers.GitProviders[0].Name)
	assert.Equal(t, "GiteaProvider", updated.Status.Providers.GitProviders[0].Kind)
	assert.True(t, updated.Status.Providers.GitProviders[0].Ready)

	// Verify Ready condition
	assert.Len(t, updated.Status.Conditions, 1)
	assert.Equal(t, "Ready", updated.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, updated.Status.Conditions[0].Status)
}
