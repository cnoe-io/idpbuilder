package k8s

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newDeployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func TestConvertYamlToObjects(t *testing.T) {
	cases := []struct {
		name          string
		schemeBuilder runtime.SchemeBuilder
		input         string
		expectErr     error
		expectObjects []client.Object
	}{{
		name:          "Single Deployment",
		schemeBuilder: appsv1.SchemeBuilder,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment1
spec:`,
		expectErr: nil,
		expectObjects: []client.Object{
			newDeployment("test-deployment1"),
		},
	}, {
		name:          "Multi Deployment",
		schemeBuilder: appsv1.SchemeBuilder,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment1
spec:
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment2
spec:`,
		expectErr: nil,
		expectObjects: []client.Object{
			newDeployment("test-deployment1"),
			newDeployment("test-deployment2"),
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			tc.schemeBuilder.AddToScheme(scheme)
			objs, err := ConvertYamlToObjects(scheme, []byte(tc.input))

			if err != tc.expectErr {
				t.Fatalf("want err: %v, got err %v", tc.expectErr, err)
			}

			if diff := cmp.Diff(tc.expectObjects, objs); diff != "" {
				t.Errorf("ConvertYamlToObjects() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
