package gitrepository

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/google/go-github/v61/github"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	gitHubTokenKey = "token"
)

type ghClient struct {
	client *github.Client
}

func (g *ghClient) getRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	r, resp, err := g.client.Repositories.Get(ctx, owner, repo)
	if resp.StatusCode == http.StatusNotFound {
		return nil, notFoundError{}
	}
	return r, err
}

func (g *ghClient) createRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	r := github.Repository{
		Name:    &repo,
		Private: github.Bool(false),
	}
	rp, _, err := g.client.Repositories.Create(ctx, owner, &r)
	if err != nil {
		return nil, err
	}
	return rp, nil
}

type gitHubProvider struct {
	client.Client
	Scheme       *runtime.Scheme
	gitHubClient gitHubClient
	config       util.CorePackageTemplateConfig
}

func (g *gitHubProvider) createRepository(ctx context.Context, repo *v1alpha1.GitRepository) (repoInfo, error) {

	r, err := g.gitHubClient.createRepo(ctx, getOrganizationName(*repo), getRepositoryName(*repo))
	if err != nil {
		return repoInfo{}, fmt.Errorf("creating repo: %w", err)
	}
	return repoInfo{
		name:                     *r.Name,
		cloneUrl:                 *r.CloneURL,
		internalGitRepositoryUrl: "",
		fullName:                 *r.FullName,
	}, nil
}

func (g *gitHubProvider) getRepository(ctx context.Context, repo *v1alpha1.GitRepository) (repoInfo, error) {

	r, err := g.gitHubClient.createRepo(ctx, getOrganizationName(*repo), getRepositoryName(*repo))
	if err != nil {
		return repoInfo{}, fmt.Errorf("creating repo: %w", err)
	}
	return repoInfo{
		name:                     *r.Name,
		cloneUrl:                 *r.CloneURL,
		internalGitRepositoryUrl: "",
		fullName:                 *r.FullName,
	}, nil
}

func (g *gitHubProvider) getProviderCredentials(ctx context.Context, repo *v1alpha1.GitRepository) (gitProviderCredentials, error) {
	var secret v1.Secret
	err := g.Client.Get(ctx, types.NamespacedName{
		Namespace: repo.Spec.SecretRef.Namespace,
		Name:      repo.Spec.SecretRef.Name,
	}, &secret)
	if err != nil {
		return gitProviderCredentials{}, err
	}

	token, ok := secret.Data[gitHubTokenKey]
	if !ok {
		return gitProviderCredentials{}, fmt.Errorf("%s key not found in secret %s in %s ns", giteaAdminUsernameKey, repo.Spec.SecretRef.Name, repo.Spec.SecretRef.Namespace)
	}

	return gitProviderCredentials{
		accessToken: string(token),
	}, nil
}

func (g *gitHubProvider) setProviderCredentials(ctx context.Context, repo *v1alpha1.GitRepository, creds gitProviderCredentials) error {
	g.gitHubClient = &ghClient{client: github.NewClient(nil).WithAuthToken(creds.accessToken)}
	return nil
}

func (g *gitHubProvider) updateRepoContent(ctx context.Context, repo *v1alpha1.GitRepository, repoInfo repoInfo, creds gitProviderCredentials) error {
	return updateRepoContent(ctx, repo, repoInfo, creds, g.Scheme, g.config)
}
