package gitrepository

import (
	"context"

	"code.gitea.io/sdk/gitea"
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
