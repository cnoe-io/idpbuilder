package custompackage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util"

	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	requeueTime = time.Second * 30
)

type Reconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
	Config   util.TemplateConfig
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pkg := v1alpha1.CustomPackage{}
	err := r.Get(ctx, req.NamespacedName, &pkg)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("reconciling custom package", "name", req.Name, "namespace", req.Namespace)
	defer r.postProcessReconcile(ctx, req, &pkg)
	result, err := r.reconcileCustomPackage(ctx, &pkg)
	if err != nil {
		r.Recorder.Event(&pkg, "Warning", "reconcile error", err.Error())
	} else {
		r.Recorder.Event(&pkg, "Normal", "reconcile success", "Successfully reconciled")
	}

	return result, err
}

func (r *Reconciler) postProcessReconcile(ctx context.Context, req ctrl.Request, pkg *v1alpha1.CustomPackage) {
	logger := log.FromContext(ctx)

	err := r.Status().Update(ctx, pkg)
	if err != nil {
		logger.Error(err, "failed updating repo status")
	}

	err = util.UpdateSyncAnnotation(ctx, r.Client, pkg)
	if err != nil {
		logger.Error(err, "failed updating repo annotation")
	}
}

// create an in-cluster repository CR, update the application spec, then apply
func (r *Reconciler) reconcileCustomPackage(ctx context.Context, resource *v1alpha1.CustomPackage) (ctrl.Result, error) {
	b, err := os.ReadFile(resource.Spec.ArgoCD.ApplicationFile)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reading file %s: %w", resource.Spec.ArgoCD.ApplicationFile, err)
	}

	var returnedRawResource []byte
	if returnedRawResource, err = util.ApplyTemplate(b, r.Config); err != nil {
		return ctrl.Result{}, err
	}

	objs, err := k8s.ConvertYamlToObjects(r.Scheme, returnedRawResource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("converting yaml to object %w", err)
	}
	if len(objs) == 0 {
		return ctrl.Result{}, fmt.Errorf("file contained 0 kubernetes objects %s", resource.Spec.ArgoCD.ApplicationFile)
	}

	app, ok := objs[0].(*argov1alpha1.Application)
	if !ok {
		return ctrl.Result{}, fmt.Errorf("object is not an PackageSpec application %s", resource.Spec.ArgoCD.ApplicationFile)
	}

	appName := app.GetName()
	if resource.Spec.Replicate {
		repoRefs := make([]v1alpha1.ObjectRef, 0, 1)
		synced := true
		if app.Spec.HasMultipleSources() {
			for j := range app.Spec.Sources {
				s := &app.Spec.Sources[j]
				res, repo, sErr := r.reconcileArgocdSource(ctx, resource, appName, resource.Spec.ArgoCD.ApplicationFile, s.RepoURL)
				if sErr != nil {
					return res, sErr
				}
				if repo != nil {
					if synced {
						synced = repo.Status.InternalGitRepositoryUrl != ""
					}
					s.RepoURL = repo.Status.InternalGitRepositoryUrl
					repoRefs = append(repoRefs, v1alpha1.ObjectRef{
						Namespace: repo.Namespace,
						Name:      repo.Name,
						UID:       string(repo.ObjectMeta.UID),
					})
				}
			}
		} else {
			s := app.Spec.Source
			res, repo, sErr := r.reconcileArgocdSource(ctx, resource, appName, resource.Spec.ArgoCD.ApplicationFile, s.RepoURL)
			if sErr != nil {
				return res, sErr
			}
			if repo != nil {
				synced = repo.Status.InternalGitRepositoryUrl != ""
				s.RepoURL = repo.Status.InternalGitRepositoryUrl
				repoRefs = append(repoRefs, v1alpha1.ObjectRef{
					Namespace: repo.Namespace,
					Name:      repo.Name,
					UID:       string(repo.ObjectMeta.UID),
				})
			}
		}
		resource.Status.GitRepositoryRefs = repoRefs
		resource.Status.Synced = synced
	}

	foundAppObj := argov1alpha1.Application{}
	err = r.Client.Get(ctx, client.ObjectKeyFromObject(app), &foundAppObj)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(ctx, app)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("creating %s app CR: %w", appName, err)
			}

			return ctrl.Result{RequeueAfter: requeueTime}, nil
		}
		return ctrl.Result{}, fmt.Errorf("getting argocd application object: %w", err)
	}

	foundAppObj.Spec = app.Spec
	foundAppObj.ObjectMeta.Annotations = app.Annotations
	foundAppObj.ObjectMeta.Labels = app.Labels
	err = r.Client.Update(ctx, &foundAppObj)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("updating argocd application object %s: %w", appName, err)
	}
	return ctrl.Result{RequeueAfter: requeueTime}, nil
}

func (r *Reconciler) reconcileArgocdSource(ctx context.Context, resource *v1alpha1.CustomPackage, appName, pkgDir, repoURL string) (ctrl.Result, *v1alpha1.GitRepository, error) {
	logger := log.FromContext(ctx)

	process, absPath, err := isCNOEDirectory(pkgDir, repoURL)
	if err != nil {
		logger.Error(err, "processing argocd app source", "dir", pkgDir, "repoURL", repoURL)
		return ctrl.Result{}, nil, err
	}
	if !process {
		return ctrl.Result{}, nil, nil
	}

	repo, err := r.reconcileGitRepo(ctx, resource, repoName(appName, absPath), absPath)
	if err != nil {
		return ctrl.Result{}, nil, err
	}

	return ctrl.Result{}, repo, nil
}

func (r *Reconciler) reconcileGitRepo(ctx context.Context, resource *v1alpha1.CustomPackage, repoName, absPath string) (*v1alpha1.GitRepository, error) {
	repo := &v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoName,
			Namespace: resource.Namespace,
		},
	}

	cliStartTime, _ := util.GetCLIStartTimeAnnotationValue(resource.ObjectMeta.Annotations)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, repo, func() error {
		if err := controllerutil.SetControllerReference(resource, repo, r.Scheme); err != nil {
			return err
		}

		if repo.ObjectMeta.Annotations == nil {
			repo.ObjectMeta.Annotations = make(map[string]string)
		}
		util.SetCLIStartTimeAnnotationValue(repo.ObjectMeta.Annotations, cliStartTime)

		repo.Spec = v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Type: "local",
				Path: absPath,
			},
			GitURL:         resource.Spec.GitServerURL,
			InternalGitURL: resource.Spec.InternalGitServeURL,
			SecretRef:      resource.Spec.GitServerAuthSecretRef,
		}

		return nil
	})
	// it's possible for an application to specify the same directory multiple times in the spec.
	// if there is a repository already created for this package, no further action is necessary.
	if !errors.IsAlreadyExists(err) {
		return repo, err
	}

	return repo, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CustomPackage{}).
		Complete(r)
}

func isCNOEDirectory(fPath, repoURL string) (bool, string, error) {
	if strings.HasPrefix(repoURL, "cnoe://") {
		parentDir := filepath.Dir(fPath)
		relativePath := strings.TrimPrefix(repoURL, "cnoe://")
		absPath, err := filepath.Abs(filepath.Join(parentDir, relativePath))
		if err != nil {
			return false, "", err
		}

		f, err := os.Stat(absPath)
		if err != nil {
			return false, "", err
		}
		if !f.IsDir() {
			return false, "", fmt.Errorf("path not a directory: %s", absPath)
		}
		return true, absPath, err
	}
	return false, "", nil
}

func repoName(appName, dir string) string {
	return fmt.Sprintf("%s-%s", appName, filepath.Base(dir))
}
