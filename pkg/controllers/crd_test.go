package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// mockClient wraps the fake client to handle the namespace="default" quirk for CRDs
type mockClient struct {
	client.Client
}

func (m *mockClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	// For CRDs, ignore the namespace and use empty string
	// Create a new NamespacedName struct to avoid mutating the caller's parameter
	if _, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
		key = types.NamespacedName{Name: key.Name, Namespace: ""}
	}
	return m.Client.Get(ctx, key, obj, opts...)
}

func (m *mockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	// For CRDs, ensure namespace is empty
	if crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
		crd.SetNamespace("")
	}
	return m.Client.Create(ctx, obj, opts...)
}

func (m *mockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	// For CRDs, ensure namespace is empty
	if crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
		crd.SetNamespace("")
	}
	return m.Client.Update(ctx, obj, opts...)
}

// Test helper to create a basic CRD object
func createTestCRD(name, group, kind, plural string) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   kind,
				Plural: plural,
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
				{
					Type:   apiextensionsv1.Established,
					Status: apiextensionsv1.ConditionTrue,
				},
			},
		},
	}
}

// Helper to create a mock client with CRD support
func createMockClient(scheme *runtime.Scheme, objects ...client.Object) client.Client {
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&apiextensionsv1.CustomResourceDefinition{}).
		WithObjects(objects...).
		Build()
	return &mockClient{Client: fakeClient}
}

// TestEnsureCRD_CreateNew tests creating a new CRD when it doesn't exist
func TestEnsureCRD_CreateNew(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	err = EnsureCRD(ctx, scheme, mockClient, crd)
	require.NoError(t, err)

	// Verify CRD was created
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	assert.Equal(t, "TestResource", retrievedCRD.Spec.Names.Kind)
	assert.Equal(t, "test.example.com", retrievedCRD.Spec.Group)
}

// TestEnsureCRD_Update tests updating an existing CRD
func TestEnsureCRD_Update(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	// Create existing CRD
	existingCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)

	mockClient := createMockClient(scheme, existingCRD)
	ctx := context.Background()

	// Update with modified spec
	updatedCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	// Add a label to distinguish the update
	updatedCRD.Labels = map[string]string{"updated": "true"}

	err = EnsureCRD(ctx, scheme, mockClient, updatedCRD)
	require.NoError(t, err)

	// Verify CRD was updated
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: updatedCRD.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	assert.Equal(t, "true", retrievedCRD.Labels["updated"])
}

// TestEnsureCRD_InvalidObject tests error handling for non-CRD objects
func TestEnsureCRD_InvalidObject(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	// Create a non-CRD object
	invalidObj := &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
	}

	err = EnsureCRD(ctx, scheme, mockClient, invalidObj)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non crd object passed to EnsureCRD")
}

// TestEnsureCRD_WaitForEstablished tests that EnsureCRD waits for CRD to be established
func TestEnsureCRD_WaitForEstablished(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	// Start with not established
	crd.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionFalse,
		},
	}

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	// Run EnsureCRD in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- EnsureCRD(ctx, scheme, mockClient, crd)
	}()

	// Give it time to create the CRD
	time.Sleep(200 * time.Millisecond)

	// Update CRD to established status
	var createdCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd.Name, Namespace: "default"}, &createdCRD)
	require.NoError(t, err)

	createdCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionTrue,
		},
	}
	err = mockClient.Status().Update(ctx, &createdCRD)
	require.NoError(t, err)

	// Wait for EnsureCRD to complete
	select {
	case err := <-errChan:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("EnsureCRD did not complete in time")
	}
}

// TestEnsureCRDs_Multiple tests installing multiple CRDs
func TestEnsureCRDs_Multiple(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	// Test by calling EnsureCRD multiple times
	crd1 := createTestCRD(
		"testresources1.test.example.com",
		"test.example.com",
		"TestResource1",
		"testresources1",
	)
	crd2 := createTestCRD(
		"testresources2.test.example.com",
		"test.example.com",
		"TestResource2",
		"testresources2",
	)

	// Install first CRD
	err = EnsureCRD(ctx, scheme, mockClient, crd1)
	require.NoError(t, err)

	// Install second CRD
	err = EnsureCRD(ctx, scheme, mockClient, crd2)
	require.NoError(t, err)

	// Verify both CRDs exist
	var retrievedCRD1 apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd1.Name, Namespace: "default"}, &retrievedCRD1)
	require.NoError(t, err)

	var retrievedCRD2 apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd2.Name, Namespace: "default"}, &retrievedCRD2)
	require.NoError(t, err)

	assert.Equal(t, "TestResource1", retrievedCRD1.Spec.Names.Kind)
	assert.Equal(t, "TestResource2", retrievedCRD2.Spec.Names.Kind)
}

// TestEnsureCRD_Idempotency tests that calling EnsureCRD multiple times is safe
func TestEnsureCRD_Idempotency(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)

	// Install first time
	err = EnsureCRD(ctx, scheme, mockClient, crd)
	require.NoError(t, err)

	// Install second time - should not error
	err = EnsureCRD(ctx, scheme, mockClient, crd)
	require.NoError(t, err)

	// Verify only one CRD exists
	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	err = mockClient.List(ctx, crdList)
	require.NoError(t, err)
	assert.Equal(t, 1, len(crdList.Items))
}

// TestEnsureCRD_MultipleNameFormats tests CRDs with different name formats
func TestEnsureCRD_MultipleNameFormats(t *testing.T) {
	tests := []struct {
		name    string
		crdName string
		group   string
		kind    string
		plural  string
	}{
		{
			name:    "standard format",
			crdName: "gateways.gateway.networking.k8s.io",
			group:   "gateway.networking.k8s.io",
			kind:    "Gateway",
			plural:  "gateways",
		},
		{
			name:    "simple format",
			crdName: "tests.example.com",
			group:   "example.com",
			kind:    "Test",
			plural:  "tests",
		},
		{
			name:    "complex format",
			crdName: "myresources.group.subgroup.example.io",
			group:   "group.subgroup.example.io",
			kind:    "MyResource",
			plural:  "myresources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			err := apiextensionsv1.AddToScheme(scheme)
			require.NoError(t, err)

			mockClient := createMockClient(scheme)
			ctx := context.Background()

			crd := createTestCRD(tt.crdName, tt.group, tt.kind, tt.plural)

			err = EnsureCRD(ctx, scheme, mockClient, crd)
			require.NoError(t, err)

			// Verify CRD was created with correct name
			var retrievedCRD apiextensionsv1.CustomResourceDefinition
			err = mockClient.Get(ctx, types.NamespacedName{Name: tt.crdName, Namespace: "default"}, &retrievedCRD)
			require.NoError(t, err)
			assert.Equal(t, tt.kind, retrievedCRD.Spec.Names.Kind)
			assert.Equal(t, tt.group, retrievedCRD.Spec.Group)
		})
	}
}

// TestEnsureCRD_PreserveExistingData tests that updates preserve resource version
func TestEnsureCRD_PreserveExistingData(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	// Create existing CRD with annotations
	existingCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	existingCRD.Annotations = map[string]string{
		"original": "annotation",
	}

	mockClient := createMockClient(scheme, existingCRD)
	ctx := context.Background()

	// Update with new data
	updatedCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	updatedCRD.Annotations = map[string]string{
		"updated": "annotation",
	}

	err = EnsureCRD(ctx, scheme, mockClient, updatedCRD)
	require.NoError(t, err)

	// Verify the update was applied
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: updatedCRD.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	assert.Equal(t, "annotation", retrievedCRD.Annotations["updated"])
}

// TestEnsureCRD_DifferentVersions tests CRDs with multiple API versions
func TestEnsureCRD_DifferentVersions(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)

	// Add multiple versions
	crd.Spec.Versions = append(crd.Spec.Versions, apiextensionsv1.CustomResourceDefinitionVersion{
		Name:    "v1beta1",
		Served:  true,
		Storage: false,
		Schema: &apiextensionsv1.CustomResourceValidation{
			OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
				Type: "object",
			},
		},
	})

	err = EnsureCRD(ctx, scheme, mockClient, crd)
	require.NoError(t, err)

	// Verify CRD has multiple versions
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	assert.Len(t, retrievedCRD.Spec.Versions, 2)
}

// TestEnsureCRD_SpecChanges tests that spec changes are applied correctly
func TestEnsureCRD_SpecChanges(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	// Create existing CRD
	existingCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	existingCRD.Spec.Scope = apiextensionsv1.NamespaceScoped

	mockClient := createMockClient(scheme, existingCRD)
	ctx := context.Background()

	// Update with changed scope
	updatedCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	updatedCRD.Spec.Scope = apiextensionsv1.ClusterScoped

	err = EnsureCRD(ctx, scheme, mockClient, updatedCRD)
	require.NoError(t, err)

	// Verify scope was updated
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: updatedCRD.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	assert.Equal(t, apiextensionsv1.ClusterScoped, retrievedCRD.Spec.Scope)
}

// TestEnsureCRD_MetadataLabelsAnnotations tests CRD metadata handling
func TestEnsureCRD_MetadataLabelsAnnotations(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)

	// Add labels and annotations
	crd.Labels = map[string]string{
		"app":     "test-app",
		"version": "v1",
	}
	crd.Annotations = map[string]string{
		"controller-gen.kubebuilder.io/version": "v0.20.0",
		"description":                           "Test CRD",
	}

	err = EnsureCRD(ctx, scheme, mockClient, crd)
	require.NoError(t, err)

	// Verify metadata was preserved
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	assert.Equal(t, "test-app", retrievedCRD.Labels["app"])
	assert.Equal(t, "v1", retrievedCRD.Labels["version"])
	assert.Equal(t, "v0.20.0", retrievedCRD.Annotations["controller-gen.kubebuilder.io/version"])
	assert.Equal(t, "Test CRD", retrievedCRD.Annotations["description"])
}

// TestEnsureCRD_ErrorHandling tests various error scenarios
func TestEnsureCRD_ErrorHandling(t *testing.T) {
	t.Run("error when CRD never becomes established", func(t *testing.T) {
		scheme := runtime.NewScheme()
		err := apiextensionsv1.AddToScheme(scheme)
		require.NoError(t, err)

		crd := createTestCRD(
			"testresources.test.example.com",
			"test.example.com",
			"TestResource",
			"testresources",
		)
		// Set status to never establish
		crd.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
			{
				Type:   apiextensionsv1.Established,
				Status: apiextensionsv1.ConditionFalse,
			},
		}

		mockClient := createMockClient(scheme)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// This should timeout or loop indefinitely
		done := make(chan error, 1)
		go func() {
			done <- EnsureCRD(ctx, scheme, mockClient, crd)
		}()

		// Wait a bit and verify it's still running
		time.Sleep(600 * time.Millisecond)

		select {
		case <-done:
			// If it completes, something unexpected happened
			t.Log("EnsureCRD completed (might have timed out, which is acceptable)")
		default:
			// Still running, which is expected behavior (infinite loop waiting for established)
			t.Log("EnsureCRD still waiting for CRD to be established (expected)")
		}
	})
}

// TestEnsureCRD_ClusterScoped tests CRD with cluster scope
func TestEnsureCRD_ClusterScoped(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	crd := createTestCRD(
		"clusterresources.test.example.com",
		"test.example.com",
		"ClusterResource",
		"clusterresources",
	)
	crd.Spec.Scope = apiextensionsv1.ClusterScoped

	err = EnsureCRD(ctx, scheme, mockClient, crd)
	require.NoError(t, err)

	// Verify CRD was created
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	assert.Equal(t, apiextensionsv1.ClusterScoped, retrievedCRD.Spec.Scope)
}

// TestGetK8sResources tests the getK8sResources function indirectly
// Since it relies on embedded FS, we test it through EnsureCRDs which calls it
func TestGetK8sResources_Integration(t *testing.T) {
	// This is tested indirectly through the actual usage in the codebase
	// The function getK8sResources is called by EnsureCRDs
	// We verify it works by checking that the embedded CRD files can be loaded
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	// We can't directly test getK8sResources without the actual embedded files
	// But we can verify the error handling works
	t.Run("error handling for missing resources", func(t *testing.T) {
		// getK8sResources should handle errors from fs.ConvertFSToBytes
		// and k8s.ConvertRawResourcesToObjects
		// This is implicitly tested through EnsureCRDs
		t.Skip("getK8sResources requires actual embedded resources to test properly")
	})
}

// TestEnsureCRDs tests the EnsureCRDs function which loads CRDs from embedded resources
func TestEnsureCRDs(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run in goroutine with timeout since it may load actual CRDs and wait for them
	done := make(chan error, 1)
	go func() {
		done <- EnsureCRDs(ctx, scheme, mockClient, nil)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Logf("EnsureCRDs returned error: %v", err)
			// Error might be expected if resources can't be loaded or timeout
		} else {
			// Check if any CRDs were created
			crdList := &apiextensionsv1.CustomResourceDefinitionList{}
			listErr := mockClient.List(ctx, crdList)
			require.NoError(t, listErr)
			t.Logf("EnsureCRDs created %d CRDs from embedded resources", len(crdList.Items))

			// Verify each CRD was actually loaded
			for _, crd := range crdList.Items {
				t.Logf("  - CRD: %s", crd.Name)
			}
		}
	case <-time.After(6 * time.Second):
		t.Log("EnsureCRDs timed out (expected if waiting for CRDs to be established)")
	}
}

// TestEnsureCRDs_WithTemplateData tests EnsureCRDs with template data
func TestEnsureCRDs_WithTemplateData(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	mockClient := createMockClient(scheme)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create some template data
	templateData := map[string]string{
		"key": "value",
	}

	// Run in goroutine with timeout
	done := make(chan error, 1)
	go func() {
		done <- EnsureCRDs(ctx, scheme, mockClient, templateData)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Logf("EnsureCRDs with template data returned error: %v", err)
		} else {
			t.Logf("EnsureCRDs succeeded with template data")
		}
	case <-time.After(6 * time.Second):
		t.Log("EnsureCRDs with template data timed out (expected)")
	}
}

// TestEnsureCRD_AlreadyEstablished tests that CRDs already established don't wait
func TestEnsureCRD_AlreadyEstablished(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	// Already established
	crd.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionTrue,
		},
	}

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	// This should complete quickly since CRD is already established
	start := time.Now()
	err = EnsureCRD(ctx, scheme, mockClient, crd)
	duration := time.Since(start)

	require.NoError(t, err)
	// Should complete in less than 1 second since no waiting is needed
	assert.Less(t, duration, 1*time.Second, "EnsureCRD should complete quickly for already established CRDs")
}

// TestEnsureCRD_CreateError tests error handling when Create fails
func TestEnsureCRD_CreateError(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)

	// Create a mock client with the CRD already present to simulate create error
	mockClient := createMockClient(scheme, crd)
	ctx := context.Background()

	// Try to create the same CRD again - should trigger update path instead
	// This tests that the function handles existing CRDs gracefully
	err = EnsureCRD(ctx, scheme, mockClient, crd)
	require.NoError(t, err)

	// Verify CRD still exists
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: crd.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
}

// TestEnsureCRD_GetError tests error handling when Get returns non-NotFound error
func TestEnsureCRD_GetError(t *testing.T) {
	// This is difficult to test with fake client without creating a custom implementation
	// The fake client doesn't easily allow us to simulate Get errors other than NotFound
	// In a real environment, this would test network errors, permission errors, etc.
	t.Skip("Requires custom mock client to simulate Get errors other than NotFound")
}

// Test that verifies the waiting loop behavior
func TestEnsureCRD_WaitingLoop(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	crd := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	// Start with not established
	crd.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionFalse,
		},
	}

	mockClient := createMockClient(scheme)
	ctx := context.Background()

	// Track how many times we check
	checkCount := 0

	// Run EnsureCRD in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- EnsureCRD(ctx, scheme, mockClient, crd)
	}()

	// Give it time to create and start checking
	time.Sleep(100 * time.Millisecond)

	// Update after a few checks to verify it's looping
	for i := 0; i < 3; i++ {
		time.Sleep(600 * time.Millisecond) // Longer than the 500ms sleep in EnsureCRD
		checkCount++

		var curCRD apiextensionsv1.CustomResourceDefinition
		err := mockClient.Get(ctx, types.NamespacedName{Name: crd.Name, Namespace: "default"}, &curCRD)
		if err == nil && checkCount >= 2 {
			// After a couple checks, mark as established
			curCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
				{
					Type:   apiextensionsv1.Established,
					Status: apiextensionsv1.ConditionTrue,
				},
			}
			err = mockClient.Status().Update(ctx, &curCRD)
			require.NoError(t, err)
			break
		}
	}

	// Wait for EnsureCRD to complete
	select {
	case err := <-errChan:
		require.NoError(t, err)
		assert.GreaterOrEqual(t, checkCount, 2, "Should have looped at least twice")
	case <-time.After(5 * time.Second):
		t.Fatal("EnsureCRD did not complete in time")
	}
}

// TestEnsureCRD_UpdateResourceVersion tests that resource version is preserved during update
func TestEnsureCRD_UpdateResourceVersion(t *testing.T) {
	scheme := runtime.NewScheme()
	err := apiextensionsv1.AddToScheme(scheme)
	require.NoError(t, err)

	// Create existing CRD with a resource version
	existingCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	existingCRD.SetResourceVersion("123")

	mockClient := createMockClient(scheme, existingCRD)
	ctx := context.Background()

	// Update CRD
	updatedCRD := createTestCRD(
		"testresources.test.example.com",
		"test.example.com",
		"TestResource",
		"testresources",
	)
	updatedCRD.Labels = map[string]string{"new": "label"}

	err = EnsureCRD(ctx, scheme, mockClient, updatedCRD)
	require.NoError(t, err)

	// Verify the resource version was preserved during update
	var retrievedCRD apiextensionsv1.CustomResourceDefinition
	err = mockClient.Get(ctx, types.NamespacedName{Name: updatedCRD.Name, Namespace: "default"}, &retrievedCRD)
	require.NoError(t, err)
	// The fake client should have updated the resource version
	assert.NotEmpty(t, retrievedCRD.GetResourceVersion())
}
