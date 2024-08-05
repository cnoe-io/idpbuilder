package localbuild

import (
	"context"
	"testing"

	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeKubeClient struct {
	mock.Mock
	client.Client
}

func (f *fakeKubeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := f.Called(ctx, list, opts)
	return args.Error(0)
}

func (f *fakeKubeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := f.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

type testCase struct {
	err         error
	listApps    []argov1alpha1.Application
	annotations []map[string]string
}

func TestGetRawInstallResources(t *testing.T) {
	e := EmbeddedInstallation{
		resourceFS:   installArgoFS,
		resourcePath: "resources/argo",
	}
	resources, err := util.ConvertFSToBytes(e.resourceFS, e.resourcePath,
		util.CorePackageTemplateConfig{
			Protocol:       "",
			Host:           "",
			Port:           "",
			UsePathRouting: false,
		},
	)
	if err != nil {
		t.Fatalf("GetRawInstallResources() error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("GetRawInstallResources() resources len != 2, got %d", len(resources))
	}

	resourcePrefix := "# UCP ARGO INSTALL RESOURCES\n"
	checkPrefix := resources[1][0:len(resourcePrefix)]
	if resourcePrefix != string(checkPrefix) {
		t.Fatalf("GetRawInstallResources() expected 1 resource with prefix %q, got %q", resourcePrefix, checkPrefix)
	}
}

func TestGetK8sInstallResources(t *testing.T) {
	e := EmbeddedInstallation{
		resourceFS:   installArgoFS,
		resourcePath: "resources/argo",
	}
	objs, err := e.installResources(k8s.GetScheme(), util.CorePackageTemplateConfig{
		Protocol:       "",
		Host:           "",
		Port:           "",
		UsePathRouting: false,
	})
	if err != nil {
		t.Fatalf("GetK8sInstallResources() error: %v", err)
	}

	if len(objs) != 58 {
		t.Fatalf("Expected 58 Argo Install Resources, got: %d", len(objs))
	}
}

func TestArgoCDAppAnnotation(t *testing.T) {
	ctx := context.Background()

	cases := []testCase{
		{
			err: nil,
			listApps: []argov1alpha1.Application{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       argov1alpha1.ApplicationSchemaGroupVersionKind.Kind,
						APIVersion: argov1alpha1.ApplicationSchemaGroupVersionKind.GroupVersion().String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nil-annotation",
						Namespace: "argocd",
					},
				},
			},
			annotations: []map[string]string{
				{
					argoCDApplicationAnnotationKeyRefresh: argoCDApplicationAnnotationValueRefreshNormal,
				},
			},
		},
		{
			err: nil,
			listApps: []argov1alpha1.Application{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       argov1alpha1.ApplicationSchemaGroupVersionKind.Kind,
						APIVersion: argov1alpha1.ApplicationSchemaGroupVersionKind.GroupVersion().String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "existing-annotation",
						Namespace: "argocd",
						Annotations: map[string]string{
							"test": "value",
						},
					},
				},
			},
			annotations: []map[string]string{
				{
					"test":                                "value",
					argoCDApplicationAnnotationKeyRefresh: argoCDApplicationAnnotationValueRefreshNormal,
				},
			},
		},
		{
			err: nil,
			listApps: []argov1alpha1.Application{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       argov1alpha1.ApplicationSchemaGroupVersionKind.Kind,
						APIVersion: argov1alpha1.ApplicationSchemaGroupVersionKind.GroupVersion().String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owned-by-appset",
						Namespace: "argocd",
						Annotations: map[string]string{
							"test": "value",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind: "ApplicationSet",
							},
						},
					},
				},
			},
			annotations: nil,
		},
		{
			err: nil,
			listApps: []argov1alpha1.Application{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       argov1alpha1.ApplicationSchemaGroupVersionKind.Kind,
						APIVersion: argov1alpha1.ApplicationSchemaGroupVersionKind.GroupVersion().String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owned-by-non-appset",
						Namespace: "argocd",
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind: "Something",
							},
						},
					},
				},
			},
			annotations: []map[string]string{
				{
					argoCDApplicationAnnotationKeyRefresh: argoCDApplicationAnnotationValueRefreshNormal,
				},
			},
		},
	}

	for i := range cases {
		c := cases[i]
		fClient := new(fakeKubeClient)
		fClient.On("List", ctx, mock.Anything, []client.ListOption{client.InNamespace(globals.ArgoCDNamespace)}).
			Run(func(args mock.Arguments) {
				apps := args.Get(1).(*argov1alpha1.ApplicationList)
				apps.Items = c.listApps
			}).Return(c.err)
		for j := range c.annotations {
			app := c.listApps[j]
			u := makeUnstructured(app.Name, app.Namespace, app.GroupVersionKind(), c.annotations[j])
			fClient.On("Patch", ctx, u, client.Apply, []client.PatchOption{client.FieldOwner(v1alpha1.FieldManager)}).Return(nil)
		}
		rec := LocalbuildReconciler{
			Client: fClient,
		}
		err := rec.requestArgoCDAppRefresh(ctx)
		fClient.AssertExpectations(t)
		assert.NoError(t, err)
	}
}

func makeUnstructured(name, namespace string, gvk schema.GroupVersionKind, annotations map[string]string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetAnnotations(annotations)
	u.SetName(name)
	u.SetNamespace(namespace)
	u.SetGroupVersionKind(gvk)
	return u
}
