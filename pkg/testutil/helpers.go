package testutil

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/printer/types"
	"github.com/go-git/go-billy/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FakeKubeClient provides a mock Kubernetes client for testing
type FakeKubeClient struct {
	mock.Mock
	client.Client
}

func (f *FakeKubeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := f.Called(ctx, key, obj, opts)
	return args.Error(0)
}

func (f *FakeKubeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := f.Called(ctx, list, opts)
	return args.Error(0)
}

// Selector creates a label selector for package-based secrets
func Selector(pkgName string) labels.Selector {
	r1, _ := labels.NewRequirement(v1alpha1.CLISecretLabelKey, selection.Equals, []string{v1alpha1.CLISecretLabelValue})
	r2, _ := labels.NewRequirement(v1alpha1.PackageNameLabelKey, selection.Equals, []string{pkgName})
	return labels.NewSelector().Add(*r1).Add(*r2)
}

// SecretDataToSecret converts a types.Secret to a v1.Secret for testing
func SecretDataToSecret(data types.Secret) v1.Secret {
	d := make(map[string][]byte)
	if data.IsCore {
		d["username"] = []byte(data.Username)
		d["password"] = []byte(data.Password)
		d["token"] = []byte(data.Token)
	} else {
		for k := range data.Data {
			d[k] = []byte(data.Data[k])
		}
	}
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: data.Name, Namespace: data.Namespace},
		Data:       d,
	}
}

// TestCopiedFiles recursively tests that files were copied correctly between filesystems
func TestCopiedFiles(t *testing.T, src, dst billy.Filesystem, srcStartPath, dstStartPath string) {
	files, err := src.ReadDir(srcStartPath)
	assert.Nil(t, err)

	for i := range files {
		file := files[i]
		if file.Mode().IsRegular() {
			// Use local ReadWorktreeFile function (will be imported from util package where needed)
			srcB, err := readWorktreeFile(src, filepath.Join(srcStartPath, file.Name()))
			assert.Nil(t, err)

			dstB, err := readWorktreeFile(dst, filepath.Join(dstStartPath, file.Name()))
			assert.Nil(t, err)
			assert.Equal(t, srcB, dstB)
		}
		if file.IsDir() {
			TestCopiedFiles(t, src, dst, filepath.Join(srcStartPath, file.Name()), filepath.Join(dstStartPath, file.Name()))
		}
	}
}

// readWorktreeFile is a local helper that reads a file from a filesystem
func readWorktreeFile(wt billy.Filesystem, path string) ([]byte, error) {
	f, err := wt.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

// MakeUnstructured creates an unstructured object for testing
func MakeUnstructured(name, namespace string, gvk schema.GroupVersionKind, annotations map[string]string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(namespace)
	if annotations != nil {
		obj.SetAnnotations(annotations)
	}
	return obj
}

// NewDeployment creates a test deployment object
func NewDeployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// Contains checks if a string contains a substring (simple helper for tests)
func Contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && 
		   (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || 
		    s[len(s)-len(substr):] == substr || 
		    func() bool {
			    for i := 0; i <= len(s)-len(substr); i++ {
				    if s[i:i+len(substr)] == substr {
					    return true
				    }
			    }
			    return false
		    }()))
}