package gitrepository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/localbuild"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	DefaultBranchName     = "main"
	giteaAdminUsernameKey = "username"
	giteaAdminPasswordKey = "password"
	requeueTime           = time.Second * 30
	gitCommitAuthorName   = "git-reconciler"
	gitCommitAuthorEmail  = "invalid@cnoe.io"
)

type GiteaClientFunc func(url string, options ...gitea.ClientOption) (GiteaClient, error)

func NewGiteaClient(url string, options ...gitea.ClientOption) (GiteaClient, error) {
	return gitea.NewClient(url, options...)
}

type RepositoryReconciler struct {
	client.Client
	GiteaClientFunc GiteaClientFunc
	Recorder        record.EventRecorder
	Scheme          *runtime.Scheme
}

func getRepositoryName(repo v1alpha1.GitRepository) string {
	return fmt.Sprintf("%s-%s", repo.Namespace, repo.Name)
}

func getOrganizationName(repo v1alpha1.GitRepository) string {
	return "giteaAdmin"
}

func (r *RepositoryReconciler) getCredentials(ctx context.Context, repo *v1alpha1.GitRepository) (string, string, error) {
	var secret v1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: repo.Spec.SecretRef.Namespace,
		Name:      repo.Spec.SecretRef.Name,
	}, &secret)
	if err != nil {
		return "", "", err
	}

	username, ok := secret.Data[giteaAdminUsernameKey]
	if !ok {
		return "", "", fmt.Errorf("%s key not found in secret %s in %s ns", giteaAdminUsernameKey, repo.Spec.SecretRef.Name, repo.Spec.SecretRef.Namespace)
	}
	password, ok := secret.Data[giteaAdminPasswordKey]
	if !ok {
		return "", "", fmt.Errorf("%s key not found in secret %s in %s ns", giteaAdminPasswordKey, repo.Spec.SecretRef.Name, repo.Spec.SecretRef.Namespace)
	}
	return string(username), string(password), nil
}

func (r *RepositoryReconciler) getBasicAuth(ctx context.Context, repo *v1alpha1.GitRepository) (http.BasicAuth, error) {
	u, p, err := r.getCredentials(ctx, repo)
	if err != nil {
		return http.BasicAuth{}, err
	}
	return http.BasicAuth{
		Username: u,
		Password: p,
	}, nil
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var gitRepo v1alpha1.GitRepository
	err := r.Get(ctx, req.NamespacedName, &gitRepo)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.shouldProcess(gitRepo) {
		return ctrl.Result{Requeue: false}, nil
	}

	logger.Info("reconciling GitRepository", "name", req.Name, "namespace", req.Namespace)
	defer r.postProcessReconcile(ctx, req, &gitRepo)

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
}

func (r *RepositoryReconciler) reconcileGitRepo(ctx context.Context, repo *v1alpha1.GitRepository) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconciling", "name", repo.Name, "dir", repo.Spec.Source)
	giteaClient, err := r.GiteaClientFunc(repo.Spec.GitURL)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, fmt.Errorf("failed to get gitea client: %w", err)
	}

	user, pass, err := r.getCredentials(ctx, repo)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, fmt.Errorf("failed to get gitea credentials: %w", err)
	}

	giteaClient.SetBasicAuth(user, pass)
	giteaClient.SetContext(ctx)

	giteaRepo, err := reconcileRepo(giteaClient, repo)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, fmt.Errorf("failed to create or update repo %w", err)
	}
	repo.Status.ExternalGitRepositoryUrl = giteaRepo.CloneURL

	err = r.reconcileRepoContent(ctx, repo, giteaRepo)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, fmt.Errorf("failed to reconcile repo content %w", err)
	}
	repo.Status.Synced = true
	return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, nil
}

func (r *RepositoryReconciler) reconcileRepoContent(ctx context.Context, repo *v1alpha1.GitRepository, giteaRepo *gitea.Repository) error {
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-%s", repo.Name, repo.Namespace))
	defer os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("creating temporary directory: %w", err)
	}

	clonedRepo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:        giteaRepo.CloneURL,
		NoCheckout: true,
	})
	if err != nil {
		return fmt.Errorf("cloning repo: %w", err)
	}

	err = writeRepoContents(repo, tempDir)
	if err != nil {
		return err
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

	auth, err := r.getBasicAuth(ctx, repo)
	if err != nil {
		return fmt.Errorf("getting basic auth: %w", err)
	}
	err = clonedRepo.Push(&git.PushOptions{
		Auth: &auth,
	})
	if err != nil {
		return fmt.Errorf("pushing to git: %w", err)
	}

	repo.Status.LatestCommit.Hash = commit.String()

	return nil
}

func reconcileRepo(giteaClient GiteaClient, repo *v1alpha1.GitRepository) (*gitea.Repository, error) {
	resp, repoResp, err := giteaClient.GetRepo(getOrganizationName(*repo), getRepositoryName(*repo))
	if err != nil {
		if repoResp.StatusCode == 404 {
			createResp, _, CErr := giteaClient.CreateRepo(gitea.CreateRepoOption{
				Name:        getRepositoryName(*repo),
				Description: fmt.Sprintf("created by Git Repository controller for %s in %s namespace", repo.Name, repo.Namespace),
				// we should reconsider this when targeting non-local clusters.
				Private:       false,
				DefaultBranch: DefaultBranchName,
				AutoInit:      true,
			})
			if CErr != nil {
				return &gitea.Repository{}, fmt.Errorf("failed to create git repository: %w", CErr)
			}
			repo.Status.ExternalGitRepositoryUrl = createResp.CloneURL
			return createResp, nil
		}
	}
	return resp, nil
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
	// embedded fs does not change
	if repo.Spec.Source.Type == "embedded" && repo.Status.Synced {
		return false
	}
	return true
}

func writeRepoContents(repo *v1alpha1.GitRepository, dstPath string) error {
	if repo.Spec.Source.EmbeddedAppName != "" {
		resources, err := localbuild.GetEmbeddedRawInstallResources(repo.Spec.Source.EmbeddedAppName)
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
