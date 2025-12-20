//go:build integration
// +build integration

package gatewayprovider

import (
	"context"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// TestNginxGatewayIntegration tests NginxGateway with a real Kubernetes cluster
// This test requires:
// 1. A running Kubernetes cluster
// 2. NginxGateway CRD installed
// 3. NginxGateway controller running
//
// Run with: go test -tags=integration -v -run TestNginxGatewayIntegration
func TestNginxGatewayIntegration(t *testing.T) {
	// Get kubeconfig
	cfg, err := config.GetConfig()
	require.NoError(t, err, "Failed to get kubeconfig")

	// Create client
	k8sClient, err := client.New(cfg, client.Options{})
	require.NoError(t, err, "Failed to create Kubernetes client")

	ctx := context.Background()
	testNamespace := "nginx-gateway-test"

	// Cleanup function
	defer func() {
		t.Log("Cleaning up test resources...")
		// Delete NginxGateway
		gateway := &v1alpha2.NginxGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-integration-nginx",
				Namespace: testNamespace,
			},
		}
		_ = k8sClient.Delete(ctx, gateway)

		// Delete test namespace
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		_ = k8sClient.Delete(ctx, ns)
		t.Log("Cleanup complete")
	}()

	// Create test namespace
	testNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	err = k8sClient.Create(ctx, testNS)
	if err != nil {
		// Namespace might already exist
		t.Logf("Namespace creation warning (may already exist): %v", err)
	}

	// Create NginxGateway
	nginxGateway := &v1alpha2.NginxGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-integration-nginx",
			Namespace: testNamespace,
		},
		Spec: v1alpha2.NginxGatewaySpec{
			Namespace: "ingress-nginx-test",
			Version:   "1.13.0",
			IngressClass: v1alpha2.NginxIngressClass{
				Name:      "nginx-test",
				IsDefault: false, // Don't make it default for test
			},
		},
	}

	t.Log("Creating NginxGateway resource...")
	err = k8sClient.Create(ctx, nginxGateway)
	require.NoError(t, err, "Failed to create NginxGateway")

	// Wait for NginxGateway to be reconciled
	t.Log("Waiting for NginxGateway to be reconciled...")
	err = wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
		gateway := &v1alpha2.NginxGateway{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      nginxGateway.Name,
			Namespace: nginxGateway.Namespace,
		}, gateway)
		if err != nil {
			t.Logf("Error getting NginxGateway: %v", err)
			return false, nil
		}

		t.Logf("NginxGateway status: Phase=%s, Installed=%v", gateway.Status.Phase, gateway.Status.Installed)

		// Check if phase is Ready
		if gateway.Status.Phase == "Ready" {
			return true, nil
		}

		return false, nil
	})
	require.NoError(t, err, "NginxGateway did not become Ready in time")

	// Verify NginxGateway status
	t.Log("Verifying NginxGateway status...")
	gateway := &v1alpha2.NginxGateway{}
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      nginxGateway.Name,
		Namespace: nginxGateway.Namespace,
	}, gateway)
	require.NoError(t, err)

	// Verify duck-typed fields are populated
	assert.Equal(t, "Ready", gateway.Status.Phase, "Phase should be Ready")
	assert.True(t, gateway.Status.Installed, "Installed should be true")
	assert.Equal(t, "nginx-test", gateway.Status.IngressClassName, "IngressClassName should match spec")
	assert.NotEmpty(t, gateway.Status.InternalEndpoint, "InternalEndpoint should be set")
	assert.Equal(t, "1.13.0", gateway.Status.Version, "Version should match spec")

	// Verify conditions
	readyCondition := false
	for _, cond := range gateway.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == metav1.ConditionTrue {
			readyCondition = true
			break
		}
	}
	assert.True(t, readyCondition, "Ready condition should be True")

	// Verify nginx resources are created
	t.Log("Verifying nginx deployment exists...")
	deployment := &appsv1.Deployment{}
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "ingress-nginx-controller",
		Namespace: "ingress-nginx-test",
	}, deployment)
	assert.NoError(t, err, "Nginx deployment should exist")

	// Verify deployment is ready
	if err == nil {
		t.Logf("Deployment status: Replicas=%d, ReadyReplicas=%d", deployment.Status.Replicas, deployment.Status.ReadyReplicas)
		assert.Greater(t, deployment.Status.Replicas, int32(0), "Deployment should have replicas")
		assert.Equal(t, deployment.Status.Replicas, deployment.Status.ReadyReplicas, "All replicas should be ready")
	}

	// Verify service exists
	t.Log("Verifying nginx service exists...")
	service := &corev1.Service{}
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "ingress-nginx-controller",
		Namespace: "ingress-nginx-test",
	}, service)
	assert.NoError(t, err, "Nginx service should exist")

	t.Log("Integration test completed successfully!")
}

// TestNginxGatewayE2E tests end-to-end creation and deletion
func TestNginxGatewayE2E(t *testing.T) {
	cfg, err := config.GetConfig()
	require.NoError(t, err)

	k8sClient, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	ctx := context.Background()
	testNamespace := "nginx-e2e-test"

	// Create test namespace
	testNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_ = k8sClient.Create(ctx, testNS)

	defer func() {
		// Cleanup
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		_ = k8sClient.Delete(ctx, ns)
	}()

	// Create NginxGateway
	nginxGateway := &v1alpha2.NginxGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-nginx",
			Namespace: testNamespace,
		},
		Spec: v1alpha2.NginxGatewaySpec{
			Namespace: "ingress-nginx-e2e",
			Version:   "1.13.0",
		},
	}

	t.Log("Creating NginxGateway for E2E test...")
	err = k8sClient.Create(ctx, nginxGateway)
	require.NoError(t, err)

	// Wait for Ready
	t.Log("Waiting for NginxGateway to become Ready...")
	err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		gateway := &v1alpha2.NginxGateway{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      nginxGateway.Name,
			Namespace: nginxGateway.Namespace,
		}, gateway)
		if err != nil {
			return false, nil
		}
		return gateway.Status.Phase == "Ready", nil
	})
	require.NoError(t, err, "NginxGateway should become Ready")

	// Delete NginxGateway
	t.Log("Deleting NginxGateway...")
	err = k8sClient.Delete(ctx, nginxGateway)
	require.NoError(t, err)

	// Wait for deletion
	t.Log("Waiting for NginxGateway to be deleted...")
	err = wait.PollImmediate(2*time.Second, 1*time.Minute, func() (bool, error) {
		gateway := &v1alpha2.NginxGateway{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      nginxGateway.Name,
			Namespace: nginxGateway.Namespace,
		}, gateway)
		// Should be not found
		return err != nil, nil
	})
	assert.NoError(t, err, "NginxGateway should be deleted")

	t.Log("E2E test completed successfully!")
}
