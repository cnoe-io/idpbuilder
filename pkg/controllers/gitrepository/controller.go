package gitrepository

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/go-git/go-git/v5"
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
	return "giteaAdmin"
}

func GetGitProvider(ctx context.Context, repo *v1alpha1.GitRepository, kubeClient client.Client, scheme *runtime.Scheme, tmplConfig util.CorePackageTemplateConfig) (gitProvider, error) {
	switch repo.Spec.Provider.Name {
	case v1alpha1.GitProviderGitea:
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c := &http.Client{Transport: tr}
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
			Client: kubeClient,
			Scheme: scheme,
			config: tmplConfig,
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
	if !r.shouldProcess(gitRepo) {
		return ctrl.Result{Requeue: false}, nil
	}

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

	err = provider.updateRepoContent(ctx, repo, providerRepo, creds)
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

func (r *RepositoryReconciler) shouldProcess(repo v1alpha1.GitRepository) bool {
	if repo.Spec.Source.Type == "local" && !filepath.IsAbs(repo.Spec.Source.Path) {
		return false
	}
	return true
}

func updateRepoContent(ctx context.Context, repo *v1alpha1.GitRepository, repoInfo repoInfo, creds gitProviderCredentials, scheme *runtime.Scheme, tmplConfig util.CorePackageTemplateConfig) error {
	logger := log.FromContext(ctx)

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-%s", repo.Name, repo.Namespace))
	defer os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("creating temporary directory: %w", err)
	}

	auth, err := getBasicAuth(creds)
	if err != nil {
		return fmt.Errorf("getting basic auth: %w", err)
	}

	cloneOptions := &git.CloneOptions{
		Auth:            &auth,
		URL:             repoInfo.cloneUrl,
		NoCheckout:      true,
		InsecureSkipTLS: true,
	}
	clonedRepo, err := git.PlainCloneContext(ctx, tempDir, false, cloneOptions)
	if err != nil {
		// if we cannot clone with gitea's configured url, then we fallback to using the url provided in spec.
		logger.V(1).Info("failed cloning with returned clone URL. Falling back to default url.", "err", err)

		cloneOptions.URL = fmt.Sprintf("%s/%s.git", repo.Spec.Provider.GitURL, repoInfo.fullName)
		c, retErr := git.PlainCloneContext(ctx, tempDir, false, cloneOptions)
		if retErr != nil {
			return fmt.Errorf("cloning repo with fall back url: %w", retErr)
		}
		clonedRepo = c
	}

	err = writeRepoContents(repo, tempDir, tmplConfig, scheme)
	if err != nil {
		return fmt.Errorf("writing repo contents: %w", err)
	}

	tree, err := clonedRepo.Worktree()
	if err != nil {
		return fmt.Errorf("getting git worktree: %w", err)
	}

	err = tree.AddGlob("*")
	if err != nil {
		return fmt.Errorf("adding git files: %w", err)
	}

	status, err := tree.Status()
	if err != nil {
		return fmt.Errorf("getting git status: %w", err)
	}

	if status.IsClean() {
		h, _ := clonedRepo.Head()
		repo.Status.LatestCommit.Hash = h.Hash().String()
		return nil
	}

	commit, err := tree.Commit(fmt.Sprintf("updated from %s", repo.Spec.Source.Path), &git.CommitOptions{
		All:               true,
		AllowEmptyCommits: false,
		Author: &object.Signature{
			Name:  gitCommitAuthorName,
			Email: gitCommitAuthorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("committing git files: %w", err)
	}

	err = clonedRepo.Push(&git.PushOptions{
		Auth:            &auth,
		InsecureSkipTLS: true,
	})
	if err != nil {
		return fmt.Errorf("pushing to git: %w", err)
	}

	repo.Status.LatestCommit.Hash = commit.String()
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
