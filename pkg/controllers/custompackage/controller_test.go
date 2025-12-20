package custompackage

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPackagePriority(t *testing.T) {
	t.Run("valid priority annotation", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "5",
				},
			},
		}
		priority, err := getPackagePriority(pkg)
		assert.NoError(t, err)
		assert.Equal(t, 5, priority)
	})

	t.Run("missing annotations", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{},
		}
		_, err := getPackagePriority(pkg)
		assert.Error(t, err)
	})

	t.Run("missing priority annotation", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"other": "value",
				},
			},
		}
		_, err := getPackagePriority(pkg)
		assert.Error(t, err)
	})

	t.Run("invalid priority format", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "invalid",
				},
			},
		}
		_, err := getPackagePriority(pkg)
		assert.Error(t, err)
	})

	t.Run("zero priority", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "0",
				},
			},
		}
		priority, err := getPackagePriority(pkg)
		assert.NoError(t, err)
		assert.Equal(t, 0, priority)
	})

	t.Run("large priority value", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "1000",
				},
			},
		}
		priority, err := getPackagePriority(pkg)
		assert.NoError(t, err)
		assert.Equal(t, 1000, priority)
	})
}
