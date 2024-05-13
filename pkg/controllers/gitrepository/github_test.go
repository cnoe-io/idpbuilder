package gitrepository

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeGH struct {
	mock.Mock
}

func (f *fakeGH) getRepo(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	args := f.Called(ctx, owner, repo)
	return args.Get(0).(*github.Repository), args.Get(1).(*github.Response), args.Error(2)
}

func (f *fakeGH) createRepo(ctx context.Context, owner string, req *github.Repository) (*github.Repository, *github.Response, error) {
	args := f.Called(ctx, owner, req)
	return args.Get(0).(*github.Repository), args.Get(1).(*github.Response), args.Error(2)
}

func (f *fakeGH) setToken(token string) error {
	return nil
}

func newResponse(r http.Response) *github.Response {
	response := &github.Response{Response: &r}
	return response
}

type fakeKubeClient struct {
	mock.Mock
	client.Client
}

func (f *fakeKubeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := f.Called(ctx, key, obj, opts)
	return args.Error(0)
}

func TestGitHubCreateRepository(t *testing.T) {
	fakeGH := new(fakeGH)
	ctx := context.Background()
	gh := gitHubProvider{
		Client:       &fakeClient{},
		gitHubClient: fakeGH,
	}
	repoExpected := repoInfo{
		name:                     "repo1",
		cloneUrl:                 "",
		internalGitRepositoryUrl: "",
		fullName:                 "owner/test-test",
	}
	resource := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Path: "ac",
				Type: "local",
			},
			Provider: v1alpha1.Provider{
				Name:             "github",
				OrganizationName: "owner",
			},
		},
	}

	expectedInput := &github.Repository{
		Name:    github.String(getRepositoryName(resource)),
		Private: github.Bool(true),
	}

	fakeGH.On("createRepo", ctx, "owner", expectedInput).Return(
		&github.Repository{
			Name:     &repoExpected.name,
			CloneURL: &repoExpected.cloneUrl,
			FullName: &repoExpected.fullName,
		},
		newResponse(http.Response{StatusCode: http.StatusOK}),
		nil,
	)

	resp, err := gh.createRepository(ctx, &resource)
	assert.Nil(t, err)
	assert.Equal(t, repoExpected, resp)
	fakeGH.AssertExpectations(t)
}

func TestGitHubGetProviderCredentials(t *testing.T) {
	fakeK8sClient := new(fakeKubeClient)
	ctx := context.Background()
	gh := gitHubProvider{
		Client: fakeK8sClient,
	}

	resource := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.GitRepositorySpec{
			SecretRef: v1alpha1.SecretReference{
				Name:      "test",
				Namespace: "testNS",
			},
		},
	}
	inputSecret := &v1.Secret{}
	fakeK8sClient.On("Get", ctx, types.NamespacedName{
		Namespace: "testNS",
		Name:      "test",
	}, inputSecret, []client.GetOption(nil)).Run(func(args mock.Arguments) {
		sec := args.Get(2).(*v1.Secret)
		sec.Data = make(map[string][]byte, 1)
		sec.Data[gitHubTokenKey] = []byte("token")
	}).Return(nil)

	creds, err := gh.getProviderCredentials(ctx, &resource)
	assert.Nil(t, err)
	assert.Equal(t, creds.accessToken, "token")
	fakeK8sClient.AssertExpectations(t)

}

func TestGitHubGetRepository(t *testing.T) {
	fakeGH := new(fakeGH)
	ctx := context.Background()
	gh := gitHubProvider{
		Client:       &fakeClient{},
		gitHubClient: fakeGH,
	}

	repoExpected := repoInfo{
		name:                     "repo1",
		cloneUrl:                 "",
		internalGitRepositoryUrl: "",
		fullName:                 "owner/test-test",
	}

	fakeGetRepo := fakeGH.On("getRepo", ctx, "owner", "test-test").Return(
		&github.Repository{
			Name:     &repoExpected.name,
			CloneURL: &repoExpected.cloneUrl,
			FullName: &repoExpected.fullName,
		},
		newResponse(http.Response{StatusCode: http.StatusOK}),
		nil,
	)

	resource := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Path: "ac",
				Type: "local",
			},
			Provider: v1alpha1.Provider{
				Name:             "github",
				OrganizationName: "owner",
			},
		},
	}

	resp, err := gh.getRepository(ctx, &resource)
	assert.Nil(t, err)
	assert.Equal(t, repoExpected, resp)
	fakeGH.AssertExpectations(t)

	fakeGetRepo.Unset()
	fakeGH.On("getRepo", ctx, "owner", "test-test").Return(
		&github.Repository{},
		newResponse(http.Response{StatusCode: http.StatusNotFound}),
		fmt.Errorf("some error"),
	)

	resp, err = gh.getRepository(ctx, &resource)
	assert.Equal(t, notFoundError{}, err)
	assert.Equal(t, repoInfo{}, resp)
	fakeGH.AssertExpectations(t)
}
