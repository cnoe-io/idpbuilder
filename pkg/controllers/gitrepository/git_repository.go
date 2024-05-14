package gitrepository

import (
	"context"

	"code.gitea.io/sdk/gitea"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/google/go-github/v61/github"
)

type GiteaClient interface {
	CreateAccessToken(option gitea.CreateAccessTokenOption) (*gitea.AccessToken, *gitea.Response, error)
	CreateOrg(opt gitea.CreateOrgOption) (*gitea.Organization, *gitea.Response, error)
	CreateRepo(opt gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error)
	DeleteOrg(orgname string) (*gitea.Response, error)
	DeleteRepo(owner, repo string) (*gitea.Response, error)
	GetOrg(orgname string) (*gitea.Organization, *gitea.Response, error)
	GetRepo(owner, reponame string) (*gitea.Repository, *gitea.Response, error)
	SetBasicAuth(username, password string)
	SetContext(ctx context.Context)
}

type gitHubClient interface {
	getRepo(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	createRepo(ctx context.Context, owner string, req *github.Repository) (*github.Repository, *github.Response, error)
	setToken(token string) error
}

type repoInfo struct {
	name                     string
	cloneUrl                 string
	internalGitRepositoryUrl string
	fullName                 string
}

type gitProviderCredentials struct {
	username    string
	password    string
	accessToken string
}

type gitProvider interface {
	createRepository(ctx context.Context, repo *v1alpha1.GitRepository) (repoInfo, error)
	getProviderCredentials(ctx context.Context, repo *v1alpha1.GitRepository) (gitProviderCredentials, error)
	getRepository(ctx context.Context, repo *v1alpha1.GitRepository) (repoInfo, error)
	setProviderCredentials(ctx context.Context, repo *v1alpha1.GitRepository, creds gitProviderCredentials) error
	updateRepoContent(ctx context.Context, repo *v1alpha1.GitRepository, repoInfo repoInfo, creds gitProviderCredentials, tmpDir string, repoMap *util.RepoMap) error
}
