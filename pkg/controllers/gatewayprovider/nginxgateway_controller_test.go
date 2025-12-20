package gatewayprovider

import (
	"context"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNginxGatewayReconciler_isNginxReady(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	tests := []struct {
		name        string
		deployment  *appsv1.Deployment
		expectReady bool
		expectError bool
	}{
		{
			name:        "deployment not found",
			deployment:  nil,
			expectReady: false,
			expectError: false,
		},
		{
			name: "deployment ready",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nginxControllerDeployment,
					Namespace: "ingress-nginx",
				},
				Status: appsv1.DeploymentStatus{
					Replicas:          1,
					AvailableReplicas: 1,
				},
			},
			expectReady: true,
			expectError: false,
		},
		{
			name: "deployment not ready",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nginxControllerDeployment,
					Namespace: "ingress-nginx",
				},
				Status: appsv1.DeploymentStatus{
					Replicas:          1,
					AvailableReplicas: 0,
				},
			},
			expectReady: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objs []client.Object
			if tt.deployment != nil {
				objs = append(objs, tt.deployment)
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(objs...).
				Build()

			reconciler := &NginxGatewayReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			gateway := &v1alpha2.NginxGateway{
				Spec: v1alpha2.NginxGatewaySpec{
					Namespace: "ingress-nginx",
				},
			}

			ready, err := reconciler.isNginxReady(context.Background(), gateway)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectReady, ready)
		})
	}
}

func TestNginxGatewayReconciler_getControllerStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	tests := []struct {
		name                string
		deployment          *appsv1.Deployment
		expectReplicas      int32
		expectReadyReplicas int32
		expectError         bool
	}{
		{
			name:                "deployment not found",
			deployment:          nil,
			expectReplicas:      0,
			expectReadyReplicas: 0,
			expectError:         true,
		},
		{
			name: "deployment with replicas",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nginxControllerDeployment,
					Namespace: "ingress-nginx",
				},
				Status: appsv1.DeploymentStatus{
					Replicas:      2,
					ReadyReplicas: 1,
				},
			},
			expectReplicas:      2,
			expectReadyReplicas: 1,
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objs []client.Object
			if tt.deployment != nil {
				objs = append(objs, tt.deployment)
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(objs...).
				Build()

			reconciler := &NginxGatewayReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			gateway := &v1alpha2.NginxGateway{
				Spec: v1alpha2.NginxGatewaySpec{
					Namespace: "ingress-nginx",
				},
			}

			replicas, readyReplicas, err := reconciler.getControllerStatus(context.Background(), gateway)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectReplicas, replicas)
				assert.Equal(t, tt.expectReadyReplicas, readyReplicas)
			}
		})
	}
}
