package localbuild

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/resources/localbuild"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultArgoCDProjectName string = "default"
)

type LocalbuildReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	CancelFunc     context.CancelFunc
	ExitOnSync     bool
	shouldShutdown bool
}

type subReconciler func(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error)

func (r *LocalbuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling", "resource", req.NamespacedName)

	var localBuild v1alpha1.Localbuild
	if err := r.Get(ctx, req.NamespacedName, &localBuild); err != nil {
		log.Error(err, "unable to fetch Resource")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Make sure we post process
	defer r.postProcessReconcile(ctx, req, &localBuild)

	// respecting order of installation matters as there are hard dependencies
	subReconcilers := []subReconciler{
		r.ReconcileProjectNamespace,
		r.ReconcileNginx,
		r.ReconcileArgo,
		r.ReconcileGitea,
		r.ReconcileArgoAppsWithGitea,
	}

	for _, sub := range subReconcilers {
		result, err := sub(ctx, req, &localBuild)
		if err != nil || result.Requeue || result.RequeueAfter != 0 {
			return result, err
		}
	}

	return ctrl.Result{}, nil
}

// Responsible to updating ObservedGeneration in status
func (r *LocalbuildReconciler) postProcessReconcile(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) {
	log := log.FromContext(ctx)

	resource.Status.ObservedGeneration = resource.GetGeneration()
	if err := r.Status().Update(ctx, resource); err != nil {
		log.Error(err, "Failed to update resource status after reconcile")
	}

	log.Info("Checking if we should shutdown")
	if r.shouldShutdown {
		log.Info("Shutting Down")
		r.CancelFunc()
	}
}

func (r *LocalbuildReconciler) ReconcileProjectNamespace(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	nsResource := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: globals.GetProjectNamespace(resource.Name),
		},
	}

	log.Info("Create or update namespace", "resource", nsResource)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, nsResource, func() error {
		if err := controllerutil.SetControllerReference(resource, nsResource, r.Scheme); err != nil {
			log.Error(err, "Setting controller ref on namespace resource")
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(err, "Create or update namespace resource")
	}
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocalbuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Localbuild{}).
		Complete(r)
}

func (r *LocalbuildReconciler) ReconcileArgoAppsWithGitea(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("installing bootstrap apps to ArgoCD")

	// push bootstrap app manifests to Gitea. let ArgoCD take over
	// will need a way to filter them based on user input
	bootStrapApps := []string{"argocd", "nginx", "gitea"}
	for _, n := range bootStrapApps {
		result, err := r.reconcileEmbeddedApp(ctx, n, resource)
		if err != nil {
			return result, fmt.Errorf("reconciling bootstrap apps %w", err)
		}
	}
	if resource.Spec.PackageConfigs.CustomPackageDirs != nil {
		for i := range resource.Spec.PackageConfigs.CustomPackageDirs {
			result, err := r.reconcileCustomPkg(ctx, resource, resource.Spec.PackageConfigs.CustomPackageDirs[i])
			if err != nil {
				return result, err
			}
		}
	}

	shutdown, err := r.shouldShutDown(ctx, resource)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	r.shouldShutdown = shutdown

	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *LocalbuildReconciler) reconcileEmbeddedApp(ctx context.Context, appName string, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Ensuring embedded ArgoCD Application", "name", appName)
	repo, err := r.reconcileGitRepo(ctx, resource, "embedded", appName, appName, "")

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating %s repo CR: %w", appName, err)
	}

	app := &argov1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: "argocd",
		},
	}

	if err := controllerutil.SetControllerReference(resource, app, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	err = r.Client.Get(ctx, client.ObjectKeyFromObject(app), app)
	if err != nil && errors.IsNotFound(err) {
		localbuild.SetApplicationSpec(
			app,
			repo.Status.InternalGitRepositoryUrl,
			".",
			defaultArgoCDProjectName,
			appName,
			nil,
		)
		err = r.Client.Create(ctx, app)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("creating %s app CR: %w", appName, err)
		}
	}

	localbuild.SetApplicationSpec(
		app,
		repo.Status.InternalGitRepositoryUrl,
		".",
		defaultArgoCDProjectName,
		appName,
		nil,
	)
	err = r.Client.Update(ctx, app)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("updating argoapp: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *LocalbuildReconciler) shouldShutDown(ctx context.Context, resource *v1alpha1.Localbuild) (bool, error) {
	if !r.ExitOnSync {
		return false, nil
	}

	repos := &v1alpha1.GitRepositoryList{}
	err := r.Client.List(ctx, repos, client.InNamespace(resource.Namespace))
	if err != nil {
		return false, fmt.Errorf("listing repositories %w", err)
	}
	for i := range repos.Items {
		repo := repos.Items[i]
		if !repo.Status.Synced {
			return false, nil
		}
	}

	pkgs := &v1alpha1.CustomPackageList{}
	err = r.Client.List(ctx, pkgs, client.InNamespace(resource.Namespace))
	if err != nil {
		return false, fmt.Errorf("listing custom packages %w", err)
	}

	for i := range pkgs.Items {
		pkg := pkgs.Items[i]
		if !pkg.Status.Synced {
			return false, nil
		}
	}

	return true, nil
}

func (r *LocalbuildReconciler) reconcileCustomPkg(ctx context.Context, resource *v1alpha1.Localbuild, pkgDir string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	files, err := os.ReadDir(pkgDir)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reading dir, %s: %w", pkgDir, err)
	}

	for i := range files {
		file := files[i]
		if !file.Type().IsRegular() {
			continue
		}

		filePath := filepath.Join(pkgDir, file.Name())
		b, fErr := os.ReadFile(filePath)
		if fErr != nil {
			logger.Error(fErr, "reading file", "file", filePath)
			continue
		}

		o := &unstructured.Unstructured{}
		_, gvk, fErr := scheme.Codecs.UniversalDeserializer().Decode(b, nil, o)
		if fErr != nil {
			continue
		}
		if gvk.Kind == "Application" && gvk.Group == "argoproj.io" {
			appName := o.GetName()
			appNS := o.GetNamespace()
			customPkg := &v1alpha1.CustomPackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      getCustomPackageName(file.Name(), appName),
					Namespace: globals.GetProjectNamespace(resource.Name),
				},
				Spec: v1alpha1.CustomPackageSpec{
					Replicate:           true,
					GitServerURL:        resource.Status.Gitea.ExternalURL,
					InternalGitServeURL: resource.Status.Gitea.InternalURL,
					GitServerAuthSecretRef: v1alpha1.SecretReference{
						Name:      resource.Status.Gitea.AdminUserSecretName,
						Namespace: resource.Status.Gitea.AdminUserSecretNamespace,
					},
					ArgoCD: v1alpha1.ArgoCDPackageSpec{
						ApplicationFile: filePath,
						Name:            appName,
						Namespace:       appNS,
					},
				},
				Status: v1alpha1.CustomPackageStatus{},
			}

			_, fErr = controllerutil.CreateOrUpdate(ctx, r.Client, customPkg, func() error {
				if err := controllerutil.SetControllerReference(resource, customPkg, r.Scheme); err != nil {
					return err
				}
				return nil
			})
			if fErr != nil {
				logger.Error(fErr, "failed creating custom package object", "name", appName, "namespace", appNS)
				continue
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *LocalbuildReconciler) reconcileGitRepo(ctx context.Context, resource *v1alpha1.Localbuild, repoType, repoName, embeddedName, absPath string) (*v1alpha1.GitRepository, error) {
	repo := &v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoName,
			Namespace: globals.GetProjectNamespace(resource.Name),
		},
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Type: repoType,
			},
			GitURL:         resource.Status.Gitea.ExternalURL,
			InternalGitURL: resource.Status.Gitea.InternalURL,
			SecretRef: v1alpha1.SecretReference{
				Name:      resource.Status.Gitea.AdminUserSecretName,
				Namespace: resource.Status.Gitea.AdminUserSecretNamespace,
			},
		},
	}

	if repoType == "embedded" {
		repo.Spec.Source.EmbeddedAppName = embeddedName
	} else {
		repo.Spec.Source.Path = absPath
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, repo, func() error {
		if err := controllerutil.SetControllerReference(resource, repo, r.Scheme); err != nil {
			return err
		}
		return nil
	})

	return repo, err
}

func getCustomPackageName(fileName, appName string) string {
	s := strings.Split(fileName, ".")
	return fmt.Sprintf("%s-%s", strings.ToLower(s[0]), appName)
}

func GetEmbeddedRawInstallResources(name string) ([][]byte, error) {
	switch name {
	case "argocd":
		return RawArgocdInstallResources()
	case "gitea":
		return RawGiteaInstallResources()
	case "nginx":
		return RawNginxInstallResources()
	default:
		return nil, fmt.Errorf("unsupported embedded app name %s", name)
	}
}
