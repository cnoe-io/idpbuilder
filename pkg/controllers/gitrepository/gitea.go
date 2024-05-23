package gitrepository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"code.gitea.io/sdk/gitea"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/localbuild"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	giteaAdminUsernameKey = "username"
	giteaAdminPasswordKey = "password"
)

type GiteaClientFunc func(url string, options ...gitea.ClientOption) (GiteaClient, error)

type giteaProvider struct {
	client.Client
	Scheme      *runtime.Scheme
	giteaClient GiteaClient
	config      util.CorePackageTemplateConfig
}

func (g *giteaProvider) createRepository(ctx context.Context, repo *v1alpha1.GitRepository) (repoInfo, error) {
	resp, _, err := g.giteaClient.CreateRepo(gitea.CreateRepoOption{
		Name:        getRepositoryName(*repo),
		Description: fmt.Sprintf("created by Git Repository controller for %s in %s namespace", repo.Name, repo.Namespace),
		// we should reconsider this when targeting non-local clusters.
		Private:       false,
		DefaultBranch: DefaultBranchName,
		AutoInit:      true,
	})

	if err != nil {
		return repoInfo{}, fmt.Errorf("failed to create git repository: %w", err)
	}
	return repoInfo{
		name:     resp.Name,
		fullName: resp.FullName,
		cloneUrl: resp.CloneURL,
	}, nil
}

func (g *giteaProvider) getProviderCredentials(ctx context.Context, repo *v1alpha1.GitRepository) (gitProviderCredentials, error) {
	var secret v1.Secret
	err := g.Client.Get(ctx, types.NamespacedName{
		Namespace: repo.Spec.SecretRef.Namespace,
		Name:      repo.Spec.SecretRef.Name,
	}, &secret)
	if err != nil {
		return gitProviderCredentials{}, err
	}

	username, ok := secret.Data[giteaAdminUsernameKey]
	if !ok {
		return gitProviderCredentials{}, fmt.Errorf("%s key not found in secret %s in %s ns", giteaAdminUsernameKey, repo.Spec.SecretRef.Name, repo.Spec.SecretRef.Namespace)
	}
	password, ok := secret.Data[giteaAdminPasswordKey]
	if !ok {
		return gitProviderCredentials{}, fmt.Errorf("%s key not found in secret %s in %s ns", giteaAdminPasswordKey, repo.Spec.SecretRef.Name, repo.Spec.SecretRef.Namespace)
	}
	return gitProviderCredentials{
		username: string(username),
		password: string(password),
	}, nil
}

func (g *giteaProvider) setProviderCredentials(ctx context.Context, repo *v1alpha1.GitRepository, creds gitProviderCredentials) error {
	g.giteaClient.SetBasicAuth(creds.username, creds.password)
	g.giteaClient.SetContext(ctx)
	return nil
}

func (g *giteaProvider) getRepository(ctx context.Context, repo *v1alpha1.GitRepository) (repoInfo, error) {
	resp, repoResp, err := g.giteaClient.GetRepo(getOrganizationName(*repo), getRepositoryName(*repo))
	if err != nil {
		if repoResp != nil && repoResp.StatusCode == 404 {
			return repoInfo{}, notFoundError{}
		}
		return repoInfo{}, err
	}

	return repoInfo{
		name:                     resp.Name,
		fullName:                 resp.FullName,
		cloneUrl:                 resp.CloneURL,
		internalGitRepositoryUrl: getInternalGiteaRepositoryURL(repo.Namespace, repo.Name, repo.Spec.Provider.InternalGitURL),
	}, nil
}

func (g *giteaProvider) updateRepoContent(
	ctx context.Context,
	repo *v1alpha1.GitRepository,
	repoInfo repoInfo,
	creds gitProviderCredentials,
	tmpDir string,
	repoMap *util.RepoMap,
) error {
	switch repo.Spec.Source.Type {
	case v1alpha1.SourceTypeLocal, v1alpha1.SourceTypeEmbedded:
		return reconcileLocalRepoContent(ctx, repo, repoInfo, creds, g.Scheme, g.config, tmpDir, repoMap)
	case v1alpha1.SourceTypeRemote:
		return reconcileRemoteRepoContent(ctx, repo, repoInfo, creds, tmpDir, repoMap)
	default:
		return nil
	}
}

func writeRepoContents(repo *v1alpha1.GitRepository, dstPath string, config util.CorePackageTemplateConfig, scheme *runtime.Scheme) error {
	if repo.Spec.Source.EmbeddedAppName != "" {
		resources, err := localbuild.GetEmbeddedRawInstallResources(
			repo.Spec.Source.EmbeddedAppName, config,
			v1alpha1.PackageCustomization{Name: repo.Spec.Customization.Name, FilePath: repo.Spec.Customization.FilePath}, scheme)
		if err != nil {
			return fmt.Errorf("getting embedded resource; %w", err)
		}

		for i := range resources {
			filePath := filepath.Join(dstPath, fmt.Sprintf("resource%d.yaml", i))
			err = os.WriteFile(filePath, resources[i], 0644)
			if err != nil {
				return fmt.Errorf("writing embedded resource; %w", err)
			}
		}
		return nil
	}

	err := util.CopyDirectory(repo.Spec.Source.Path, dstPath)
	if err != nil {
		return fmt.Errorf("copying files: %w", err)
	}
	return nil
}

func getBasicAuth(creds gitProviderCredentials) (githttp.BasicAuth, error) {
	b := githttp.BasicAuth{
		Username: creds.username,
		Password: creds.password,
	}
	if creds.password == "" {
		b.Password = creds.accessToken
	}
	return b, nil
}

func NewGiteaClient(url string, options ...gitea.ClientOption) (GiteaClient, error) {
	return gitea.NewClient(url, options...)
}

func getInternalGiteaRepositoryURL(namespace, name, baseUrl string) string {
	return fmt.Sprintf("%s/%s/%s-%s.git", baseUrl, v1alpha1.GiteaAdminUserName, namespace, name)
}
