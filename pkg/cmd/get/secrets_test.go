package get

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				{Name: argoCDInitialAdminSecretName, Namespace: "argocd"},
				{Name: giteaAdminSecretName, Namespace: "gitea"},
			},
		},
		{
			err:      nil,
			packages: []string{"argocd", "gitea"},
			getKeys: []client.ObjectKey{
				{Name: argoCDInitialAdminSecretName, Namespace: "argocd"},
				{Name: giteaAdminSecretName, Namespace: "gitea"},
			},
		},
		{
			err:      nil,
			packages: []string{"argocd"},
			getKeys: []client.ObjectKey{
				{Name: argoCDInitialAdminSecretName, Namespace: "argocd"},
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

		err := printPackageSecrets(ctx, io.Discard, fClient, "")
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
				{Name: argoCDInitialAdminSecretName, Namespace: "argocd"},
				{Name: giteaAdminSecretName, Namespace: "gitea"},
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
		err := printAllPackageSecrets(ctx, io.Discard, fClient, "")
		fClient.AssertExpectations(t)
		assert.Nil(t, err)
	}
}

func TestOutput(t *testing.T) {
	ctx := context.Background()
	r, _ := labels.NewRequirement(v1alpha1.CLISecretLabelKey, selection.Equals, []string{v1alpha1.CLISecretLabelValue})

	corePkgData := map[string]entity.Secret{
		argoCDInitialAdminSecretName: {
			IsCore:    true,
			Name:      argoCDInitialAdminSecretName,
			Namespace: "argocd",
			Username:  "admin",
			Password:  "abc",
		},
		giteaAdminSecretName: {
			IsCore:    true,
			Name:      giteaAdminSecretName,
			Namespace: "gitea",
			Username:  "admin",
			Password:  "abc",
		},
	}

	packageData := map[string]entity.Secret{
		"name1": {
			Name:      "name1",
			Namespace: "ns1",
			Data: map[string]string{
				"data1": "data1",
				"data2": "data2",
			},
		},
		"name2": {
			Name:      "name2",
			Namespace: "ns2",
			Data: map[string]string{
				"data1": "data1",
				"data2": "data2",
			},
		},
	}

	fClient := new(fakeKubeClient)
	opts := client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*r),
		Namespace:     "",
	}

	fClient.On("Get", ctx, client.ObjectKey{Name: argoCDInitialAdminSecretName, Namespace: "argocd"}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*v1.Secret)
		sec := secretDataToSecret(corePkgData[argoCDInitialAdminSecretName])
		*arg = sec
	}).Return(nil)
	fClient.On("Get", ctx, client.ObjectKey{Name: giteaAdminSecretName, Namespace: "gitea"}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*v1.Secret)
		sec := secretDataToSecret(corePkgData[giteaAdminSecretName])
		*arg = sec
	}).Return(nil)

	fClient.On("List", ctx, mock.Anything, []client.ListOption{&opts}).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*v1.SecretList)
		secs := make([]v1.Secret, 0, 2)
		for k := range packageData {
			s := secretDataToSecret(packageData[k])
			secs = append(secs, s)
		}
		arg.Items = secs
	}).Return(nil)

	var b []byte
	buffer := bytes.NewBuffer(b)

	err := printAllPackageSecrets(ctx, buffer, fClient, "json")
	fClient.AssertExpectations(t)
	assert.Nil(t, err)

	// verify received json data
	var received []entity.Secret
	err = json.Unmarshal(buffer.Bytes(), &received)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(received))

	for i := range received {
		rec := received[i]
		c, ok := corePkgData[rec.Name]
		if ok {
			// Set the IsCore bool field to true as the v1.Secret don't include it !
			rec.IsCore = true
			assert.Equal(t, c, rec)
			delete(corePkgData, rec.Name)
			continue
		} else {
			d, okE := packageData[rec.Name]
			if okE {
				assert.Equal(t, d, rec)
				delete(packageData, rec.Name)
				continue
			}
			t.Fatalf("found an invalid element: %s", rec.Name)
		}
	}
	assert.Equal(t, 0, len(corePkgData))
	assert.Equal(t, 0, len(packageData))
}

func secretDataToSecret(data entity.Secret) v1.Secret {
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
