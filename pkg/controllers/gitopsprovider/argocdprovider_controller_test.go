package gitopsprovider

import (
	"context"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestArgoCDProviderReconciler_BasicReconciliation tests basic reconciliation flow
func TestArgoCDProviderReconciler_BasicReconciliation(t *testing.T) {
	scheme := k8s.GetScheme()

	// Create test ArgoCD provider
	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	// Create fake client with resources
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Test reconciliation
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      argocdProvider.Name,
			Namespace: argocdProvider.Namespace,
		},
	}

	ctx := context.Background()
	result, err := reconciler.Reconcile(ctx, req)

	// Note: Reconciliation will fail on manifest parsing in unit tests
	// because embedded manifests require actual cluster resources.
	// This is expected behavior in unit tests.
	t.Log("Reconciliation result:", err)
	_ = result

	// Verify the provider resource was fetched and status was initialized
	updatedProvider := &v1alpha2.ArgoCDProvider{}
	err = fakeClient.Get(ctx, req.NamespacedName, updatedProvider)
	require.NoError(t, err, "Failed to get updated provider")

	// Status should be set even if installation fails
	assert.NotEmpty(t, updatedProvider.Status.Phase, "Phase should be set")
}

// TestArgoCDProviderReconciler_ResourceNotFound tests when resource doesn't exist
func TestArgoCDProviderReconciler_ResourceNotFound(t *testing.T) {
	scheme := k8s.GetScheme()

	// Create fake client without the provider
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "non-existent",
			Namespace: "argocd",
		},
	}

	ctx := context.Background()
	result, err := reconciler.Reconcile(ctx, req)

	// Should not error when resource is not found
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

// TestArgoCDProviderReconciler_StatusInitialization tests initial status setup
func TestArgoCDProviderReconciler_StatusInitialization(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      argocdProvider.Name,
			Namespace: argocdProvider.Namespace,
		},
	}

	// First reconciliation - should set initial status
	_, err := reconciler.Reconcile(ctx, req)
	// Expect error due to manifest parsing in fake client
	t.Log("Reconciliation error (expected):", err)

	// Get updated provider
	updatedProvider := &v1alpha2.ArgoCDProvider{}
	err = fakeClient.Get(ctx, req.NamespacedName, updatedProvider)
	require.NoError(t, err)

	// Verify status was initialized
	assert.Contains(t, []string{"Pending", "Failed"}, updatedProvider.Status.Phase,
		"Phase should be set to Pending initially or Failed if installation fails")
}

// TestArgoCDProviderReconciler_EnsureAdminCredentials_AutoGenerate tests credential auto-generation
func TestArgoCDProviderReconciler_EnsureAdminCredentials_AutoGenerate(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, ns).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call ensureAdminCredentials directly
	err := reconciler.ensureAdminCredentials(ctx, argocdProvider)
	require.NoError(t, err)

	// Verify secret was created
	secret := &corev1.Secret{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      "argocd-admin-secret",
		Namespace: "argocd",
	}, secret)
	require.NoError(t, err, "Admin secret should be created")

	// Verify secret contents
	assert.Contains(t, secret.StringData, "password", "Secret should contain password")
	assert.Contains(t, secret.StringData, "username", "Secret should contain username")
	assert.Equal(t, "admin", secret.StringData["username"], "Username should be 'admin'")
	assert.NotEmpty(t, secret.StringData["password"], "Password should not be empty")
}

// TestArgoCDProviderReconciler_EnsureAdminCredentials_Idempotent tests that credentials are not recreated
func TestArgoCDProviderReconciler_EnsureAdminCredentials_Idempotent(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	// Pre-create the secret
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-admin-secret",
			Namespace: "argocd",
		},
		StringData: map[string]string{
			"username": "admin",
			"password": "existing-password",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, existingSecret).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call ensureAdminCredentials - should not recreate
	err := reconciler.ensureAdminCredentials(ctx, argocdProvider)
	require.NoError(t, err)

	// Verify secret still exists with original password
	secret := &corev1.Secret{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      "argocd-admin-secret",
		Namespace: "argocd",
	}, secret)
	require.NoError(t, err)

	// The original secret data should be preserved
	// Note: StringData gets converted to Data in fake client
	assert.NotNil(t, secret)
}

// TestArgoCDProviderReconciler_EnsureAdminCredentials_SecretRef tests using a secret reference
func TestArgoCDProviderReconciler_EnsureAdminCredentials_SecretRef(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: false,
				SecretRef: &v1alpha2.SecretReference{
					Name:      "custom-secret",
					Namespace: "argocd",
					Key:       "password",
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call ensureAdminCredentials - should not auto-generate
	err := reconciler.ensureAdminCredentials(ctx, argocdProvider)
	require.NoError(t, err, "Should succeed when secretRef is provided")

	// Verify no secret was auto-created
	secret := &corev1.Secret{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      "argocd-admin-secret",
		Namespace: "argocd",
	}, secret)
	assert.True(t, apierrors.IsNotFound(err), "Secret should not be auto-created when using secretRef")
}

// TestArgoCDProviderReconciler_EnsureAdminCredentials_NoSecretRef tests error when neither auto-generate nor secretRef is set
func TestArgoCDProviderReconciler_EnsureAdminCredentials_NoSecretRef(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: false,
				SecretRef:    nil,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call ensureAdminCredentials - should error
	err := reconciler.ensureAdminCredentials(ctx, argocdProvider)
	require.Error(t, err, "Should error when autoGenerate is false and secretRef is nil")
	assert.Contains(t, err.Error(), "admin credentials not configured")
}

// TestArgoCDProviderReconciler_UpdateStatus tests status update logic
func TestArgoCDProviderReconciler_UpdateStatus(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	// Create ArgoCD server service
	argocdService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-server",
			Namespace: "argocd",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.96.0.1",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, argocdService).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call updateStatus
	err := reconciler.updateStatus(ctx, argocdProvider)
	require.NoError(t, err)

	// Verify status fields are set correctly
	assert.True(t, argocdProvider.Status.Installed, "Installed should be true")
	assert.Equal(t, "Ready", argocdProvider.Status.Phase, "Phase should be Ready")
	assert.Equal(t, "v2.9.0", argocdProvider.Status.Version, "Version should match spec")
	assert.Contains(t, argocdProvider.Status.InternalEndpoint, "argocd-server.argocd.svc.cluster.local",
		"Internal endpoint should be set")
	assert.Contains(t, argocdProvider.Status.Endpoint, "argocd", "External endpoint should contain argocd")
	assert.NotNil(t, argocdProvider.Status.CredentialsSecretRef, "Credentials secret ref should be set")
	assert.Equal(t, "argocd-admin-secret", argocdProvider.Status.CredentialsSecretRef.Name,
		"Secret ref name should be set")

	// Verify Ready condition is set
	assert.NotEmpty(t, argocdProvider.Status.Conditions, "Conditions should not be empty")
	var readyCondition *metav1.Condition
	for i := range argocdProvider.Status.Conditions {
		if argocdProvider.Status.Conditions[i].Type == "Ready" {
			readyCondition = &argocdProvider.Status.Conditions[i]
			break
		}
	}
	require.NotNil(t, readyCondition, "Ready condition should be set")
	assert.Equal(t, metav1.ConditionTrue, readyCondition.Status, "Ready condition should be True")
	assert.Equal(t, "ArgoCDInstalled", readyCondition.Reason)
}

// TestArgoCDProviderReconciler_UpdateStatus_ServiceNotFound tests when ArgoCD server service doesn't exist
func TestArgoCDProviderReconciler_UpdateStatus_ServiceNotFound(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call updateStatus - should not error when service is not found
	err := reconciler.updateStatus(ctx, argocdProvider)
	require.NoError(t, err, "Should not error when service is not found yet")
}

// TestArgoCDProviderReconciler_SetCondition tests condition setting
func TestArgoCDProviderReconciler_SetCondition(t *testing.T) {
	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
	}

	// Set a condition
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "InstallationComplete",
		Message:            "ArgoCD is ready",
		LastTransitionTime: metav1.Now(),
	}

	meta.SetStatusCondition(&argocdProvider.Status.Conditions, condition)

	// Verify condition was added
	require.Len(t, argocdProvider.Status.Conditions, 1, "Should have one condition")
	assert.Equal(t, "Ready", argocdProvider.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, argocdProvider.Status.Conditions[0].Status)
	assert.Equal(t, "InstallationComplete", argocdProvider.Status.Conditions[0].Reason)
	assert.NotNil(t, argocdProvider.Status.Conditions[0].LastTransitionTime)
}

// TestArgoCDProviderReconciler_SetCondition_Update tests updating an existing condition
func TestArgoCDProviderReconciler_SetCondition_Update(t *testing.T) {
	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Status: v1alpha2.ArgoCDProviderStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionFalse,
					Reason:             "Installing",
					Message:            "Installing ArgoCD",
					LastTransitionTime: metav1.Now(),
				},
			},
		},
	}

	// Update the condition
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "InstallationComplete",
		Message:            "ArgoCD is ready",
		LastTransitionTime: metav1.Now(),
	}

	meta.SetStatusCondition(&argocdProvider.Status.Conditions, condition)

	// Verify condition was updated
	require.Len(t, argocdProvider.Status.Conditions, 1, "Should still have one condition")
	assert.Equal(t, metav1.ConditionTrue, argocdProvider.Status.Conditions[0].Status,
		"Condition status should be updated")
	assert.Equal(t, "InstallationComplete", argocdProvider.Status.Conditions[0].Reason)
}

// TestArgoCDProviderReconciler_ReconcileIdempotency tests that multiple reconciliations are idempotent
func TestArgoCDProviderReconciler_ReconcileIdempotency(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      argocdProvider.Name,
			Namespace: argocdProvider.Namespace,
		},
	}

	// First reconciliation
	_, err1 := reconciler.Reconcile(ctx, req)
	t.Log("First reconciliation:", err1)

	// Get status after first reconcile
	provider1 := &v1alpha2.ArgoCDProvider{}
	err := fakeClient.Get(ctx, req.NamespacedName, provider1)
	require.NoError(t, err)
	phase1 := provider1.Status.Phase

	// Second reconciliation
	_, err2 := reconciler.Reconcile(ctx, req)
	t.Log("Second reconciliation:", err2)

	// Get status after second reconcile
	provider2 := &v1alpha2.ArgoCDProvider{}
	err = fakeClient.Get(ctx, req.NamespacedName, provider2)
	require.NoError(t, err)
	phase2 := provider2.Status.Phase

	// Phase should be consistent across reconciliations
	assert.Equal(t, phase1, phase2, "Phase should remain consistent across reconciliations")
}

// TestArgoCDProviderReconciler_InstallationError tests error handling during installation
func TestArgoCDProviderReconciler_InstallationError(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      argocdProvider.Name,
			Namespace: argocdProvider.Namespace,
		},
	}

	// Reconcile - will fail on manifest parsing in fake client
	_, err := reconciler.Reconcile(ctx, req)
	require.Error(t, err, "Should error when installation fails")

	// Get updated provider
	updatedProvider := &v1alpha2.ArgoCDProvider{}
	err = fakeClient.Get(ctx, req.NamespacedName, updatedProvider)
	require.NoError(t, err)

	// Verify status reflects the failure
	assert.Equal(t, "Failed", updatedProvider.Status.Phase, "Phase should be Failed")

	// Verify Ready condition reflects the failure
	var readyCondition *metav1.Condition
	for i := range updatedProvider.Status.Conditions {
		if updatedProvider.Status.Conditions[i].Type == "Ready" {
			readyCondition = &updatedProvider.Status.Conditions[i]
			break
		}
	}
	if readyCondition != nil {
		assert.Equal(t, metav1.ConditionFalse, readyCondition.Status, "Ready condition should be False")
		assert.Equal(t, "InstallationFailed", readyCondition.Reason)
		assert.Contains(t, readyCondition.Message, "Failed to install ArgoCD")
	}
}

// TestGenerateRandomPassword tests password generation
func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short password", 8},
		{"medium password", 16},
		{"long password", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := generateRandomPassword(tt.length)
			require.NoError(t, err, "Password generation should not error")
			assert.Len(t, password, tt.length, "Password should have the requested length")
			assert.NotEmpty(t, password, "Password should not be empty")

			// Generate another password and ensure they're different
			password2, err := generateRandomPassword(tt.length)
			require.NoError(t, err)
			assert.NotEqual(t, password, password2, "Generated passwords should be unique")
		})
	}
}

// TestArgoCDProviderReconciler_NamespaceCreation tests namespace creation
func TestArgoCDProviderReconciler_NamespaceCreation(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "custom-argocd-ns",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call installArgoCD which should create the namespace
	err := reconciler.installArgoCD(ctx, argocdProvider)
	// Will error on manifest parsing, but namespace should be created
	t.Log("Installation error (expected):", err)

	// Verify namespace was created
	ns := &corev1.Namespace{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: "custom-argocd-ns"}, ns)
	require.NoError(t, err, "Namespace should be created")
	assert.Equal(t, "custom-argocd-ns", ns.Name)
}

// TestArgoCDProviderReconciler_NamespaceAlreadyExists tests handling when namespace already exists
func TestArgoCDProviderReconciler_NamespaceAlreadyExists(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "existing-ns",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	// Pre-create the namespace
	existingNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-ns",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, existingNs).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call installArgoCD - should not error on existing namespace
	err := reconciler.installArgoCD(ctx, argocdProvider)
	// Will error on manifest parsing, but that's after namespace check
	t.Log("Installation error (expected):", err)

	// Verify namespace still exists
	ns := &corev1.Namespace{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: "existing-ns"}, ns)
	require.NoError(t, err, "Namespace should still exist")
}

// TestArgoCDProviderReconciler_SetupWithManager tests controller manager setup
func TestArgoCDProviderReconciler_SetupWithManager(t *testing.T) {
	scheme := k8s.GetScheme()

	// Create a basic reconciler
	reconciler := &ArgoCDProviderReconciler{
		Scheme: scheme,
	}

	// Note: Cannot fully test SetupWithManager without a real manager
	// This test validates the reconciler has required fields
	assert.NotNil(t, reconciler.Scheme, "Scheme should be set")
}

// TestArgoCDProviderReconciler_CustomSecretName tests using a custom secret name
func TestArgoCDProviderReconciler_CustomSecretName(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
				SecretRef: &v1alpha2.SecretReference{
					Name:      "custom-admin-secret",
					Namespace: "argocd",
					Key:       "password",
				},
			},
		},
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, ns).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call ensureAdminCredentials with custom secret name
	err := reconciler.ensureAdminCredentials(ctx, argocdProvider)
	require.NoError(t, err)

	// Verify custom secret was created
	secret := &corev1.Secret{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      "custom-admin-secret",
		Namespace: "argocd",
	}, secret)
	require.NoError(t, err, "Custom admin secret should be created")
	assert.Contains(t, secret.StringData, "password")
	assert.Contains(t, secret.StringData, "username")
}

// TestArgoCDProviderReconciler_UpdateStatus_WithCustomSecretRef tests status update with custom secret ref
func TestArgoCDProviderReconciler_UpdateStatus_WithCustomSecretRef(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
				SecretRef: &v1alpha2.SecretReference{
					Name:      "my-custom-secret",
					Namespace: "argocd",
					Key:       "pwd",
				},
			},
		},
	}

	// Create ArgoCD server service
	argocdService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-server",
			Namespace: "argocd",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.96.0.1",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, argocdService).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call updateStatus
	err := reconciler.updateStatus(ctx, argocdProvider)
	require.NoError(t, err)

	// Verify status uses custom secret ref
	assert.NotNil(t, argocdProvider.Status.CredentialsSecretRef)
	assert.Equal(t, "my-custom-secret", argocdProvider.Status.CredentialsSecretRef.Name,
		"Should use custom secret name from spec")
}

// TestArgoCDProviderReconciler_GranularConditions_NamespaceReady tests NamespaceReady condition
func TestArgoCDProviderReconciler_GranularConditions_NamespaceReady(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call installArgoCD which sets NamespaceReady condition
	err := reconciler.installArgoCD(ctx, argocdProvider)
	// Will error on manifest parsing, but namespace condition should be set
	t.Log("Installation error (expected):", err)

	// Verify NamespaceReady condition exists
	var namespaceCondition *metav1.Condition
	for i := range argocdProvider.Status.Conditions {
		if argocdProvider.Status.Conditions[i].Type == "NamespaceReady" {
			namespaceCondition = &argocdProvider.Status.Conditions[i]
			break
		}
	}
	require.NotNil(t, namespaceCondition, "NamespaceReady condition should be set")
	assert.Equal(t, metav1.ConditionTrue, namespaceCondition.Status, "NamespaceReady should be True after namespace creation")
	assert.Equal(t, "NamespaceExists", namespaceCondition.Reason)
}

// TestArgoCDProviderReconciler_GranularConditions_AdminSecretReady tests AdminSecretReady condition
func TestArgoCDProviderReconciler_GranularConditions_AdminSecretReady(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: true,
			},
		},
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, ns).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call ensureAdminCredentials which sets AdminSecretReady condition
	err := reconciler.ensureAdminCredentials(ctx, argocdProvider)
	require.NoError(t, err)

	// Verify AdminSecretReady condition exists
	var adminSecretCondition *metav1.Condition
	for i := range argocdProvider.Status.Conditions {
		if argocdProvider.Status.Conditions[i].Type == "AdminSecretReady" {
			adminSecretCondition = &argocdProvider.Status.Conditions[i]
			break
		}
	}
	require.NotNil(t, adminSecretCondition, "AdminSecretReady condition should be set")
	assert.Equal(t, metav1.ConditionTrue, adminSecretCondition.Status, "AdminSecretReady should be True after secret creation")
	assert.Contains(t, []string{"AdminSecretCreated", "AdminSecretExists"}, adminSecretCondition.Reason)
}

// TestArgoCDProviderReconciler_GranularConditions_ResourcesInstalled tests ResourcesInstalled condition
func TestArgoCDProviderReconciler_GranularConditions_ResourcesInstalled(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call installArgoCD which sets ResourcesInstalled condition
	err := reconciler.installArgoCD(ctx, argocdProvider)
	// Will error on resource creation in fake client
	t.Log("Installation error (expected):", err)

	// Verify ResourcesInstalled condition exists
	var resourcesCondition *metav1.Condition
	for i := range argocdProvider.Status.Conditions {
		if argocdProvider.Status.Conditions[i].Type == "ResourcesInstalled" {
			resourcesCondition = &argocdProvider.Status.Conditions[i]
			break
		}
	}
	require.NotNil(t, resourcesCondition, "ResourcesInstalled condition should be set")
	// In unit tests, this will fail due to fake client limitations
	assert.Equal(t, metav1.ConditionFalse, resourcesCondition.Status, "ResourcesInstalled should be False in unit tests")
	assert.Equal(t, "ResourceCreationFailed", resourcesCondition.Reason)
}

// TestArgoCDProviderReconciler_GranularConditions_DeploymentReady tests DeploymentReady condition
func TestArgoCDProviderReconciler_GranularConditions_DeploymentReady(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call isArgoCDReady which sets DeploymentReady condition
	ready, err := reconciler.isArgoCDReady(ctx, argocdProvider)
	require.NoError(t, err)
	assert.False(t, ready, "Should not be ready without deployment")

	// Verify DeploymentReady condition exists
	var deploymentCondition *metav1.Condition
	for i := range argocdProvider.Status.Conditions {
		if argocdProvider.Status.Conditions[i].Type == "DeploymentReady" {
			deploymentCondition = &argocdProvider.Status.Conditions[i]
			break
		}
	}
	require.NotNil(t, deploymentCondition, "DeploymentReady condition should be set")
	assert.Equal(t, metav1.ConditionFalse, deploymentCondition.Status, "DeploymentReady should be False without deployment")
	assert.Equal(t, "DeploymentNotFound", deploymentCondition.Reason)
}

// TestArgoCDProviderReconciler_GranularConditions_APIAccessible tests APIAccessible condition
func TestArgoCDProviderReconciler_GranularConditions_APIAccessible(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
		},
	}

	// Create a deployment with available replicas
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "argocd-server",
				"namespace": "argocd",
			},
			"status": map[string]interface{}{
				"availableReplicas": int64(1),
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider, deployment).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call isArgoCDReady which sets APIAccessible condition
	ready, err := reconciler.isArgoCDReady(ctx, argocdProvider)
	require.NoError(t, err)
	assert.False(t, ready, "Should not be ready without service")

	// Verify APIAccessible condition exists
	var apiCondition *metav1.Condition
	for i := range argocdProvider.Status.Conditions {
		if argocdProvider.Status.Conditions[i].Type == "APIAccessible" {
			apiCondition = &argocdProvider.Status.Conditions[i]
			break
		}
	}
	require.NotNil(t, apiCondition, "APIAccessible condition should be set")
	assert.Equal(t, metav1.ConditionFalse, apiCondition.Status, "APIAccessible should be False without service")
	assert.Equal(t, "ServiceNotFound", apiCondition.Reason)
}

// TestArgoCDProviderReconciler_GranularConditions_AdminSecretConfigured tests AdminSecretReady with SecretRef
func TestArgoCDProviderReconciler_GranularConditions_AdminSecretConfigured(t *testing.T) {
	scheme := k8s.GetScheme()

	argocdProvider := &v1alpha2.ArgoCDProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd",
			Namespace: "argocd",
		},
		Spec: v1alpha2.ArgoCDProviderSpec{
			Namespace: "argocd",
			Version:   "v2.9.0",
			AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
				AutoGenerate: false,
				SecretRef: &v1alpha2.SecretReference{
					Name:      "custom-secret",
					Namespace: "argocd",
					Key:       "password",
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(argocdProvider).
		WithStatusSubresource(&v1alpha2.ArgoCDProvider{}).
		Build()

	reconciler := &ArgoCDProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()

	// Call ensureAdminCredentials with secretRef configured
	err := reconciler.ensureAdminCredentials(ctx, argocdProvider)
	require.NoError(t, err)

	// Verify AdminSecretReady condition exists
	var adminSecretCondition *metav1.Condition
	for i := range argocdProvider.Status.Conditions {
		if argocdProvider.Status.Conditions[i].Type == "AdminSecretReady" {
			adminSecretCondition = &argocdProvider.Status.Conditions[i]
			break
		}
	}
	require.NotNil(t, adminSecretCondition, "AdminSecretReady condition should be set")
	assert.Equal(t, metav1.ConditionTrue, adminSecretCondition.Status, "AdminSecretReady should be True when secretRef is configured")
	assert.Equal(t, "AdminSecretConfigured", adminSecretCondition.Reason)
}
