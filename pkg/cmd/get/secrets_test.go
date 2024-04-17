package get

import (
	"context"
	"io"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeKubeClient struct {
	mock.Mock
	client.Client
}

func (f *fakeKubeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := f.Called(ctx, key, obj, opts)
	return args.Error(0)
}

func (f *fakeKubeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := f.Called(ctx, list, opts)
	return args.Error(0)
}

type cases struct {
	err               error
	packages          []string
	getKeys           []client.ObjectKey
	listLabelSelector []labels.Selector
}

func selector(pkgName string) labels.Selector {
	r1, _ := labels.NewRequirement(v1alpha1.CLISecretLabelKey, selection.Equals, []string{v1alpha1.CLISecretLabelValue})
	r2, _ := labels.NewRequirement(v1alpha1.PackageNameLabelKey, selection.Equals, []string{pkgName})
	return labels.NewSelector().Add(*r1).Add(*r2)
}

func TestPrintPackageSecrets(t *testing.T) {
	ctx := context.Background()

	cs := []cases{
		{
			err:               nil,
			packages:          []string{"abc"},
			listLabelSelector: []labels.Selector{selector("abc")},
		},
		{
			err:               nil,
			packages:          []string{"argocd", "gitea", "abc"},
			listLabelSelector: []labels.Selector{selector("abc")},
			getKeys: []client.ObjectKey{
				{Name: "argocd-initial-admin-secret", Namespace: "argocd"},
				{Name: "gitea-admin-secret", Namespace: "gitea"},
			},
		},
		{
			err:      nil,
			packages: []string{"argocd", "gitea"},
			getKeys: []client.ObjectKey{
				{Name: "argocd-initial-admin-secret", Namespace: "argocd"},
				{Name: "gitea-admin-secret", Namespace: "gitea"},
			},
		},
		{
			err:      nil,
			packages: []string{"argocd"},
			getKeys: []client.ObjectKey{
				{Name: "argocd-initial-admin-secret", Namespace: "argocd"},
			},
		},
	}

	for i := range cs {
		c := cs[i]
		fClient := new(fakeKubeClient)
		packages = c.packages

		for j := range c.listLabelSelector {
			opts := client.ListOptions{
				LabelSelector: c.listLabelSelector[j],
				Namespace:     "",
			}
			fClient.On("List", ctx, mock.Anything, []client.ListOption{&opts}).Return(c.err)
		}

		for j := range c.getKeys {
			fClient.On("Get", ctx, c.getKeys[j], mock.Anything, mock.Anything).Return(c.err)
		}

		err := printPackageSecrets(ctx, io.Discard, fClient)
		fClient.AssertExpectations(t)
		assert.Nil(t, err)
	}
}

func TestPrintAllPackageSecrets(t *testing.T) {
	ctx := context.Background()

	r, _ := labels.NewRequirement(v1alpha1.CLISecretLabelKey, selection.Equals, []string{v1alpha1.CLISecretLabelValue})

	cs := []cases{
		{
			err:               nil,
			listLabelSelector: []labels.Selector{labels.NewSelector().Add(*r)},
			getKeys: []client.ObjectKey{
				{Name: "argocd-initial-admin-secret", Namespace: "argocd"},
				{Name: "gitea-admin-secret", Namespace: "gitea"},
			},
		},
	}

	for i := range cs {
		c := cs[i]
		fClient := new(fakeKubeClient)

		for j := range c.listLabelSelector {
			opts := client.ListOptions{
				LabelSelector: c.listLabelSelector[j],
				Namespace:     "",
			}
			fClient.On("List", ctx, mock.Anything, []client.ListOption{&opts}).Return(c.err)
		}

		for j := range c.getKeys {
			fClient.On("Get", ctx, c.getKeys[j], mock.Anything, mock.Anything).Return(c.err)
		}

		err := printAllPackageSecrets(ctx, io.Discard, fClient)
		fClient.AssertExpectations(t)
		assert.Nil(t, err)
	}
}
