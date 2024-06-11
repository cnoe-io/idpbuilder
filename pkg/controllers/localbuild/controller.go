package localbuild

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	argocdapp "github.com/cnoe-io/argocd-api/api/argo/application"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/resources/localbuild"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

var (
	defaultRequeueTime = time.Second * 15
	errRequeueTime     = time.Second * 5
)

type LocalbuildReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	CancelFunc     context.CancelFunc
	ExitOnSync     bool
	shouldShutdown bool
	Config         util.CorePackageTemplateConfig
	TempDir        string
	RepoMap        *util.RepoMap
}

type subReconciler func(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error)

func (r *LocalbuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Reconciling", "resource", req.NamespacedName)

	var localBuild v1alpha1.Localbuild
	if err := r.Get(ctx, req.NamespacedName, &localBuild); err != nil {
		logger.Error(err, "unable to fetch Resource")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Make sure we post process
	defer r.postProcessReconcile(ctx, req, &localBuild)

	_, err := r.ReconcileProjectNamespace(ctx, req, &localBuild)
	if err != nil {
		return ctrl.Result{}, err
	}

	instCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error, 3)

	go r.installCorePackages(instCtx, req, &localBuild, errChan)

	select {
	case <-ctx.Done():
		return ctrl.Result{}, nil
	case instErr := <-errChan:
		if instErr != nil {
			// likely due to ingress-nginx admission hook not ready. debug log and try again.
			logger.V(1).Info("failed installing core package. likely not fatal. will try again", "error", instErr)
			return ctrl.Result{RequeueAfter: errRequeueTime}, nil
		}
	}

	logger.V(1).Info("done installing core packages. passing control to argocd")
	_, err = r.ReconcileArgoAppsWithGitea(ctx, req, &localBuild)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
}

func (r *LocalbuildReconciler) installCorePackages(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild, errChan chan error) {
	logger := log.FromContext(ctx)
	defer close(errChan)
	var wg sync.WaitGroup

	installers := map[string]subReconciler{
		"nginx":  r.ReconcileNginx,
		"argocd": r.ReconcileArgo,
		"gitea":  r.ReconcileGitea,
	}
	logger.V(1).Info("installing core packages")
	for k, v := range installers {
		wg.Add(1)
		name := k
		inst := v
		go func() {
			defer wg.Done()
			_, iErr := inst(ctx, req, resource)
			if iErr != nil {
				logger.V(1).Info("failed installing %s: %s", name, iErr)
				errChan <- fmt.Errorf("failed installing %s: %w", name, iErr)
			}
		}()
	}
	wg.Wait()
}

// Responsible to updating ObservedGeneration in status
func (r *LocalbuildReconciler) postProcessReconcile(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) {
	log := log.FromContext(ctx)

	log.Info("Checking if we should shutdown")
	if r.shouldShutdown {
		log.Info("Shutting Down")
		r.CancelFunc()
		return
	}

	resource.Status.ObservedGeneration = resource.GetGeneration()
	if err := r.Status().Update(ctx, resource); err != nil {
		log.Error(err, "Failed to update resource status after reconcile")
	}
}

func (r *LocalbuildReconciler) ReconcileProjectNamespace(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	nsResource := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: globals.GetProjectNamespace(resource.Name),
		},
	}

	logger.V(1).Info("Create or update namespace", "resource", nsResource)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, nsResource, func() error {
		if err := controllerutil.SetControllerReference(resource, nsResource, r.Scheme); err != nil {
			logger.Error(err, "Setting controller ref on namespace resource")
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Create or update namespace resource")
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

	for _, s := range resource.Spec.PackageConfigs.CustomPackageDirs {
		result, err := r.reconcileCustomPkgDir(ctx, resource, s)
		if err != nil {
			return result, err
		}
	}

	for _, s := range resource.Spec.PackageConfigs.CustomPackageUrls {
		result, err := r.reconcileCustomPkgUrl(ctx, resource, s)
		if err != nil {
			return result, err
		}
	}

	shutdown, err := r.shouldShutDown(ctx, resource)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	r.shouldShutdown = shutdown

	return ctrl.Result{}, nil
}

func (r *LocalbuildReconciler) reconcileEmbeddedApp(ctx context.Context, appName string, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.V(1).Info("Ensuring embedded ArgoCD Application", "name", appName)
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
	if err != nil && k8serrors.IsNotFound(err) {
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
	logger := log.FromContext(ctx)

	if !r.ExitOnSync {
		return false, nil
	}

	cliStartTime, err := util.GetCLIStartTimeAnnotationValue(resource.Annotations)
	if err != nil {
		return false, err
	}

	repos := &v1alpha1.GitRepositoryList{}
	err = r.Client.List(ctx, repos, client.InNamespace(resource.Namespace))
	if err != nil {
		return false, fmt.Errorf("listing repositories %w", err)
	}

	for i := range repos.Items {
		repo := repos.Items[i]

		startTimeAnnotation, gErr := util.GetCLIStartTimeAnnotationValue(repo.ObjectMeta.Annotations)
		if gErr != nil {
			// this means this repository resource is not managed by localbuild
			continue
		}

		// this object is not part of this CLI invocation
		if startTimeAnnotation != cliStartTime {
			continue
		}

		observedTime, gErr := util.GetLastObservedSyncTimeAnnotationValue(repo.ObjectMeta.Annotations)
		if gErr != nil {
			logger.Info(gErr.Error())
			return false, nil
		}

		if !repo.Status.Synced || cliStartTime != observedTime {
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
		startTimeAnnotation, gErr := util.GetCLIStartTimeAnnotationValue(pkg.ObjectMeta.Annotations)
		if gErr != nil {
			continue
		}

		if startTimeAnnotation != cliStartTime {
			return false, nil
		}

		observedTime, gErr := util.GetLastObservedSyncTimeAnnotationValue(pkg.ObjectMeta.Annotations)
		if gErr != nil {
			logger.Info(gErr.Error())
			return false, nil
		}
		if !pkg.Status.Synced || cliStartTime != observedTime {
			return false, nil
		}
	}

	return true, nil
}

func (r *LocalbuildReconciler) reconcileCustomPkg(
	ctx context.Context,
	resource *v1alpha1.Localbuild,
	b []byte,
	filePath string,
	remote *util.KustomizeRemote,
) error {
	o := &unstructured.Unstructured{}
	_, gvk, fErr := scheme.Codecs.UniversalDeserializer().Decode(b, nil, o)
	if fErr != nil {
		return fErr
	}

	if isSupportedArgoCDTypes(gvk) {
		kind := o.GetKind()
		appName := o.GetName()
		appNS := o.GetNamespace()
		customPkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getCustomPackageName(filepath.Base(filePath), appName),
				Namespace: globals.GetProjectNamespace(resource.Name),
			},
		}

		cliStartTime, _ := util.GetCLIStartTimeAnnotationValue(resource.ObjectMeta.Annotations)

		_, fErr = controllerutil.CreateOrUpdate(ctx, r.Client, customPkg, func() error {
			if err := controllerutil.SetControllerReference(resource, customPkg, r.Scheme); err != nil {
				return err
			}
			if customPkg.ObjectMeta.Annotations == nil {
				customPkg.ObjectMeta.Annotations = make(map[string]string)
			}

			util.SetCLIStartTimeAnnotationValue(customPkg.ObjectMeta.Annotations, cliStartTime)

			customPkg.Spec = v1alpha1.CustomPackageSpec{
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
					Type:            kind,
				},
			}

			if remote != nil {
				customPkg.Spec.RemoteRepository = v1alpha1.RemoteRepositorySpec{
					Url:             remote.CloneUrl(),
					Ref:             remote.Ref,
					CloneSubmodules: remote.Submodules,
					Path:            remote.Path(),
				}
			}

			return nil
		})
		return fErr
	}
	return nil
}

func (r *LocalbuildReconciler) reconcileCustomPkgUrl(ctx context.Context, resource *v1alpha1.Localbuild, pkgUrl string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	remote, err := util.NewKustomizeRemote(pkgUrl)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("parsing url, %s: %w", pkgUrl, err)
	}
	rs := v1alpha1.RemoteRepositorySpec{
		Url:             remote.CloneUrl(),
		Ref:             remote.Ref,
		CloneSubmodules: remote.Submodules,
		Path:            remote.Path(),
	}

	cloneDir := util.RepoDir(rs.Url, r.TempDir)
	st := r.RepoMap.LoadOrStore(rs.Url, cloneDir)
	st.MU.Lock()
	defer st.MU.Unlock()
	wt, _, err := util.CloneRemoteRepoToDir(ctx, rs, 1, false, cloneDir, "")
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("cloning repo, %s: %w", pkgUrl, err)
	}

	yamlFiles, err := util.GetWorktreeYamlFiles(remote.Path(), wt, false)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting yaml files from repo, %s: %w", pkgUrl, err)
	}

	for _, yamlFile := range yamlFiles {
		b, fErr := util.ReadWorktreeFile(wt, yamlFile)
		if fErr != nil {
			logger.V(1).Info("processing", "file", yamlFile, "err", fErr)
			continue
		}

		rErr := r.reconcileCustomPkg(ctx, resource, b, yamlFile, remote)
		if rErr != nil {
			logger.Error(rErr, "reconciling custom pkg", "file", yamlFile, "pkgUrl", pkgUrl)
		}
	}
	return ctrl.Result{}, nil
}

func (r *LocalbuildReconciler) reconcileCustomPkgDir(ctx context.Context, resource *v1alpha1.Localbuild, pkgDir string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	files, err := os.ReadDir(pkgDir)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reading dir, %s: %w", pkgDir, err)
	}

	for i := range files {
		file := files[i]
		if !file.Type().IsRegular() || !util.IsYamlFile(file.Name()) {
			continue
		}

		filePath := filepath.Join(pkgDir, file.Name())
		b, fErr := os.ReadFile(filePath)
		if fErr != nil {
			logger.Error(fErr, "reading file", "file", filePath)
			continue
		}

		rErr := r.reconcileCustomPkg(ctx, resource, b, filePath, nil)
		if rErr != nil {
			logger.Error(rErr, "reconciling custom pkg", "file", filePath, "pkgDir", pkgDir)
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
	}

	cliStartTime, err := util.GetCLIStartTimeAnnotationValue(resource.Annotations)
	if err != nil {
		return nil, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, repo, func() error {
		if err := controllerutil.SetControllerReference(resource, repo, r.Scheme); err != nil {
			return err
		}

		if repo.ObjectMeta.Annotations == nil {
			repo.ObjectMeta.Annotations = make(map[string]string)
		}
		util.SetCLIStartTimeAnnotationValue(repo.ObjectMeta.Annotations, cliStartTime)

		repo.Spec = v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Type: repoType,
			},
			Provider: v1alpha1.Provider{
				Name:             v1alpha1.GitProviderGitea,
				GitURL:           resource.Status.Gitea.ExternalURL,
				InternalGitURL:   resource.Status.Gitea.InternalURL,
				OrganizationName: v1alpha1.GiteaAdminUserName,
			},
			SecretRef: v1alpha1.SecretReference{
				Name:      resource.Status.Gitea.AdminUserSecretName,
				Namespace: resource.Status.Gitea.AdminUserSecretNamespace,
			},
		}

		if repoType == v1alpha1.SourceTypeEmbedded {
			repo.Spec.Source.EmbeddedAppName = embeddedName
		} else {
			repo.Spec.Source.Path = absPath
		}
		f, ok := resource.Spec.PackageConfigs.EmbeddedArgoApplications.PackageCustomization[embeddedName]
		if ok {
			repo.Spec.Customization = v1alpha1.PackageCustomization{
				Name:     embeddedName,
				FilePath: f.FilePath,
			}
		}
		return nil
	})

	return repo, err
}

func getCustomPackageName(fileName, appName string) string {
	s := strings.Split(fileName, ".")
	return fmt.Sprintf("%s-%s", strings.ToLower(s[0]), appName)
}

func isSupportedArgoCDTypes(gvk *schema.GroupVersionKind) bool {
	if gvk == nil {
		return false
	}
	return gvk.Group == argocdapp.Group && (gvk.Kind == argocdapp.ApplicationKind || gvk.Kind == argocdapp.ApplicationSetKind)
}

func GetEmbeddedRawInstallResources(name string, templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	switch name {
	case "argocd":
		return RawArgocdInstallResources(templateData, config, scheme)
	case "gitea":
		return RawGiteaInstallResources(templateData, config, scheme)
	case "nginx":
		return RawNginxInstallResources(templateData, config, scheme)
	default:
		return nil, fmt.Errorf("unsupported embedded app name %s", name)
	}
}
