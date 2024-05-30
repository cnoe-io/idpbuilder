package gitrepository

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitclient "github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	DefaultBranchName    = "main"
	requeueTime          = time.Second * 30
	gitCommitAuthorName  = "git-reconciler"
	gitCommitAuthorEmail = "idpbuilder-agent@cnoe.io"

	gitTCPTimeout = 5 * time.Second
	// timeout value for a git operation through http. clone, push, etc.
	gitHTTPTimeout = 30 * time.Second
)

func init() {
	configureGitClient()
}

type RepositoryReconciler struct {
	client.Client
	Recorder        record.EventRecorder
	Scheme          *runtime.Scheme
	Config          util.CorePackageTemplateConfig
	GitProviderFunc gitProviderFunc
	TempDir         string
	RepoMap         *util.RepoMap
}

type gitProviderFunc func(context.Context, *v1alpha1.GitRepository, client.Client, *runtime.Scheme, util.CorePackageTemplateConfig) (gitProvider, error)

type notFoundError struct{}

func (n notFoundError) Error() string {
	return fmt.Sprintf("repo not found")
}

func getRepositoryName(repo v1alpha1.GitRepository) string {
	return fmt.Sprintf("%s-%s", repo.Namespace, repo.Name)
}

func getOrganizationName(repo v1alpha1.GitRepository) string {
	return repo.Spec.Provider.OrganizationName
}

func getFallbackRepositoryURL(repo *v1alpha1.GitRepository, info repoInfo) string {
	return fmt.Sprintf("%s/%s.git", repo.Spec.Provider.GitURL, info.fullName)
}

func GetGitProvider(ctx context.Context, repo *v1alpha1.GitRepository, kubeClient client.Client, scheme *runtime.Scheme, tmplConfig util.CorePackageTemplateConfig) (gitProvider, error) {
	switch repo.Spec.Provider.Name {
	case v1alpha1.GitProviderGitea:
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   gitTCPTimeout,
				KeepAlive: 30 * time.Second, // from http.DefaultTransport
			}).DialContext,
		}
		c := &http.Client{Transport: tr, Timeout: gitHTTPTimeout}
		giteaClient, err := NewGiteaClient(repo.Spec.Provider.GitURL, gitea.SetHTTPClient(c))
		if err != nil {
			return nil, err
		}
		return &giteaProvider{
			Client:      kubeClient,
			Scheme:      scheme,
			giteaClient: giteaClient,
			config:      tmplConfig,
		}, nil
	case v1alpha1.GitProviderGitHub:
		return &gitHubProvider{
			Client:       kubeClient,
			Scheme:       scheme,
			config:       tmplConfig,
			gitHubClient: newGitHubClient(nil),
		}, nil
	}
	return nil, fmt.Errorf("invalid git provider %s ", repo.Spec.Provider.Name)
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var gitRepo v1alpha1.GitRepository
	err := r.Get(ctx, req.NamespacedName, &gitRepo)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	defer r.postProcessReconcile(ctx, req, &gitRepo)

	logger.V(1).Info("reconciling GitRepository", "name", req.Name, "namespace", req.Namespace)
	result, err := r.reconcileGitRepo(ctx, &gitRepo)
	if err != nil {
		r.Recorder.Event(&gitRepo, "Warning", "reconcile error", err.Error())
	} else {
		r.Recorder.Event(&gitRepo, "Normal", "reconcile success", "Successfully reconciled")
	}

	return result, err
}

func (r *RepositoryReconciler) postProcessReconcile(ctx context.Context, req ctrl.Request, repo *v1alpha1.GitRepository) {
	logger := log.FromContext(ctx)
	err := r.Status().Update(ctx, repo)
	if err != nil {
		logger.Error(err, "failed updating repo status")
	}

	err = util.UpdateSyncAnnotation(ctx, r.Client, repo)
	if err != nil {
		logger.Error(err, "failed updating repo annotation")
	}
}

func (r *RepositoryReconciler) reconcileGitRepo(ctx context.Context, repo *v1alpha1.GitRepository) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("reconciling", "name", repo.Name, "dir", repo.Spec.Source)
	repo.Status.Synced = false

	provider, err := r.GitProviderFunc(ctx, repo, r.Client, r.Scheme, r.Config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("initializing git provider: %w", err)
	}

	creds, err := provider.getProviderCredentials(ctx, repo)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting git provider credentials: %w", err)
	}

	err = provider.setProviderCredentials(ctx, repo, creds)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("setting git provider credentials: %w", err)
	}

	var providerRepo repoInfo
	p, err := provider.getRepository(ctx, repo)
	if err != nil {
		if errors.Is(err, notFoundError{}) {
			p, err = provider.createRepository(ctx, repo)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("creating repository: %w", err)
			}
			providerRepo = p
		} else {
			return ctrl.Result{}, fmt.Errorf("getting repository: %w", err)
		}
	} else {
		providerRepo = p
	}

	err = provider.updateRepoContent(ctx, repo, providerRepo, creds, r.TempDir, r.RepoMap)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("updating repository contents: %w", err)
	}

	repo.Status.ExternalGitRepositoryUrl = providerRepo.cloneUrl
	repo.Status.InternalGitRepositoryUrl = providerRepo.internalGitRepositoryUrl
	repo.Status.Synced = true
	return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, nil
}

func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager, notifyChan chan event.GenericEvent) error {
	// TODO: should use notifyChan to trigger reconcile when FS changes
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GitRepository{}).
		Complete(r)
}

func addAllAndCommit(path string, gitRepo *git.Repository) (plumbing.Hash, bool, error) {
	tree, err := gitRepo.Worktree()
	if err != nil {
		return plumbing.Hash{}, false, fmt.Errorf("getting git worktree: %w", err)
	}

	err = tree.AddGlob("*")
	if err != nil {
		return plumbing.Hash{}, false, fmt.Errorf("adding git files: %w", err)
	}

	status, err := tree.Status()
	if err != nil {
		return plumbing.Hash{}, false, fmt.Errorf("getting git status: %w", err)
	}

	if status.IsClean() {
		h, _ := gitRepo.Head()
		return h.Hash(), false, nil
	}

	h, err := tree.Commit(fmt.Sprintf("updated from %s", path), &git.CommitOptions{
		All:               true,
		AllowEmptyCommits: false,
		Author: &object.Signature{
			Name:  gitCommitAuthorName,
			Email: gitCommitAuthorEmail,
			When:  time.Now(),
		},
	})
	return h, true, nil
}

func pushToRemote(ctx context.Context, remoteRepo *git.Repository, creds gitProviderCredentials) error {
	auth, err := getBasicAuth(creds)
	if err != nil {
		return fmt.Errorf("getting basic auth: %w", err)
	}
	return remoteRepo.PushContext(ctx, &git.PushOptions{
		Auth:            &auth,
		InsecureSkipTLS: true,
	})
}

// add files from local fs to target repository (gitea for now)
func reconcileLocalRepoContent(ctx context.Context, repo *v1alpha1.GitRepository, tgtRepo repoInfo, creds gitProviderCredentials, scheme *runtime.Scheme, tmplConfig util.CorePackageTemplateConfig, tmpDir string, repoMap *util.RepoMap) error {
	logger := log.FromContext(ctx)
	tgtCloneDir := util.RepoDir(tgtRepo.cloneUrl, tmpDir)

	st := repoMap.LoadOrStore(tgtRepo.cloneUrl, tgtCloneDir)
	st.MU.Lock()
	defer st.MU.Unlock()

	tgtRepoSpec := v1alpha1.RemoteRepositorySpec{
		CloneSubmodules: false,
		Path:            ".",
		Url:             tgtRepo.cloneUrl,
		Ref:             "",
	}
	logger.V(1).Info("cloning repo", "repoUrl", tgtRepoSpec.Url, "fallbackUrl", getFallbackRepositoryURL(repo, tgtRepo), "cloneDir", tgtCloneDir)
	_, tgtRepository, err := util.CloneRemoteRepoToDir(ctx, tgtRepoSpec, 1, true, tgtCloneDir, getFallbackRepositoryURL(repo, tgtRepo))
	if err != nil {
		return fmt.Errorf("cloning repo %s: %w", tgtRepoSpec.Url, err)
	}

	err = writeRepoContents(repo, tgtCloneDir, tmplConfig, scheme)
	if err != nil {
		return fmt.Errorf("writing repo contents: %w", err)
	}

	hash, push, err := addAllAndCommit(repo.Spec.Source.Path, tgtRepository)
	if err != nil {
		return fmt.Errorf("add and commit %w", err)
	}

	if push {
		remoteUrl, err := util.FirstRemoteURL(tgtRepository)
		if err != nil {
			return fmt.Errorf("getting remote url %w", err)
		}

		logger.V(1).Info("pushing to remote url %s", remoteUrl)
		err = pushToRemote(ctx, tgtRepository, creds)
		if err != nil {
			return fmt.Errorf("pushing to git: %w", err)
		}

		repo.Status.LatestCommit.Hash = hash.String()
		return nil
	}

	repo.Status.LatestCommit.Hash = hash.String()
	return nil
}

// add files from another repository at specified path to target repository (gitea for now)
func reconcileRemoteRepoContent(ctx context.Context, repo *v1alpha1.GitRepository, tgtRepo repoInfo, creds gitProviderCredentials, tmpDir string, repoMap *util.RepoMap) error {
	logger := log.FromContext(ctx)
	srcRepo := repo.Spec.Source.RemoteRepository
	cloneDir := util.RepoDir(srcRepo.Url, tmpDir)

	st := repoMap.LoadOrStore(srcRepo.Url, cloneDir)
	st.MU.Lock()
	defer st.MU.Unlock()

	logger.V(1).Info("cloning repo", "repoUrl", srcRepo.Url, "fallbackUrl", "", "cloneDir", cloneDir)
	remoteWT, _, err := util.CloneRemoteRepoToDir(ctx, srcRepo, 1, false, cloneDir, "")
	if err != nil {
		return fmt.Errorf("cloning repo, %s: %w", srcRepo.Url, err)
	}

	tgtRepoSpec := v1alpha1.RemoteRepositorySpec{
		CloneSubmodules: false,
		Path:            ".",
		Url:             tgtRepo.cloneUrl,
		Ref:             "",
	}

	tgtCloneDir := util.RepoDir(tgtRepo.cloneUrl, tmpDir)
	lst := repoMap.LoadOrStore(tgtRepoSpec.Url, tgtCloneDir)

	lst.MU.Lock()
	defer lst.MU.Unlock()

	logger.V(1).Info("cloning repo", "repoUrl", tgtRepoSpec.Url, "fallbackUrl", getFallbackRepositoryURL(repo, tgtRepo), "cloneDir", tgtCloneDir)
	tgtRepoWT, tgtRepository, err := util.CloneRemoteRepoToDir(ctx, tgtRepoSpec, 1, true, tgtCloneDir, getFallbackRepositoryURL(repo, tgtRepo))
	if err != nil {
		return fmt.Errorf("cloning repo %s: %w", srcRepo.Url, err)
	}

	err = util.CopyTreeToTree(remoteWT, tgtRepoWT, fmt.Sprintf("/%s", repo.Spec.Source.Path), ".")
	if err != nil {
		return fmt.Errorf("copying contents, %s: %w", tgtRepo.cloneUrl, err)
	}

	hash, push, err := addAllAndCommit(repo.Spec.Source.Path, tgtRepository)
	if err != nil {
		return fmt.Errorf("add and commit %w", err)
	}

	if push {
		remoteUrl, err := util.FirstRemoteURL(tgtRepository)
		if err != nil {
			return fmt.Errorf("getting remote url %w", err)
		}

		logger.V(1).Info("pushing to remote url %s", remoteUrl)
		err = pushToRemote(ctx, tgtRepository, creds)
		if err != nil {
			return fmt.Errorf("pushing to git: %w", err)
		}

		repo.Status.LatestCommit.Hash = hash.String()
		return nil
	}

	repo.Status.LatestCommit.Hash = hash.String()
	return nil
}

func configureGitClient() {
	tr := http.DefaultTransport.(*http.Transport).Clone()

	tr.DialContext = (&net.Dialer{
		Timeout:   gitTCPTimeout,
		KeepAlive: 30 * time.Second, // from http.DefaultTransport
	}).DialContext

	customClient := &http.Client{
		Transport: tr,
		Timeout:   gitHTTPTimeout,
	}
	gitclient.InstallProtocol("https", githttp.NewClient(customClient))
	gitclient.InstallProtocol("http", githttp.NewClient(customClient))
}
