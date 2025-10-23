package custompackage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	argocdapplication "github.com/cnoe-io/argocd-api/api/argo/application"
	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
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
	Config   v1alpha1.BuildCustomizationSpec
	TempDir  string
	RepoMap  *util.RepoMap
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pkg := v1alpha1.CustomPackage{}
	err := r.Get(ctx, req.NamespacedName, &pkg)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(1).Info("reconciling custom package", "name", req.Name, "namespace", req.Namespace)
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

// shouldTakeOverGitRepository checks if this CustomPackage should take over an existing GitRepository.
// Returns true if this package has higher priority than the current owner.
func (r *Reconciler) shouldTakeOverGitRepository(ctx context.Context, resource *v1alpha1.CustomPackage, existingRepo *v1alpha1.GitRepository) (bool, error) {
	logger := log.FromContext(ctx)

	// Get this package's priority
	thisPriority, err := getPackagePriority(resource)
	if err != nil {
		// If no priority annotation, assume it's a legacy package - allow takeover for backward compat
		logger.V(1).Info("no priority on this package, allowing takeover", "package", resource.Name)
		return true, nil
	}

	// Check the existing repo's owner references
	for _, ownerRef := range existingRepo.GetOwnerReferences() {
		if ownerRef.Kind == "CustomPackage" {
			// Fetch the owner CustomPackage to get its priority
			ownerPkg := &v1alpha1.CustomPackage{}
			err := r.Client.Get(ctx, client.ObjectKey{
				Name:      ownerRef.Name,
				Namespace: existingRepo.Namespace,
			}, ownerPkg)

			if err != nil {
				if errors.IsNotFound(err) {
					// Owner doesn't exist anymore, we can take over
					logger.Info("GitRepository owner not found, taking over", "repo", existingRepo.Name)
					return true, nil
				}
				return false, fmt.Errorf("getting owner package: %w", err)
			}

			// Get owner's priority
			ownerPriority, err := getPackagePriority(ownerPkg)
			if err != nil {
				// Owner has no priority, we can take over
				logger.V(1).Info("GitRepository owner has no priority, taking over", "repo", existingRepo.Name)
				return true, nil
			}

			// Only take over if we have HIGHER priority (higher number wins)
			if thisPriority > ownerPriority {
				logger.Info("Taking over GitRepository from lower priority owner",
					"repo", existingRepo.Name,
					"ourPriority", thisPriority,
					"ownerPriority", ownerPriority,
					"owner", ownerPkg.Name)
				return true, nil
			} else {
				logger.Info("Not taking over GitRepository owned by higher/equal priority package",
					"repo", existingRepo.Name,
					"ourPriority", thisPriority,
					"ownerPriority", ownerPriority,
					"owner", ownerPkg.Name)
				return false, nil
			}
		}
	}

	// No CustomPackage owner found, we can take over
	logger.V(1).Info("GitRepository has no CustomPackage owner, taking over", "repo", existingRepo.Name)
	return true, nil
}

// shouldReconcile checks if this CustomPackage should proceed with reconciliation.
// It returns true only if this is the highest priority CustomPackage for the same app.
func (r *Reconciler) shouldReconcile(ctx context.Context, resource *v1alpha1.CustomPackage) (bool, error) {
	logger := log.FromContext(ctx)

	// Get this package's priority
	thisPriority, err := getPackagePriority(resource)
	if err != nil {
		// If no priority annotation, assume it's a legacy package and allow it
		logger.V(1).Info("no priority annotation found, assuming legacy package", "name", resource.Name)
		return true, nil
	}

	// List all CustomPackages in the same namespace
	pkgList := &v1alpha1.CustomPackageList{}
	err = r.Client.List(ctx, pkgList, client.InNamespace(resource.Namespace))
	if err != nil {
		return false, fmt.Errorf("listing custom packages: %w", err)
	}

	// Check if any other CustomPackage has the same app name with higher priority
	for i := range pkgList.Items {
		pkg := &pkgList.Items[i]

		// Skip self
		if pkg.Name == resource.Name {
			continue
		}

		// Skip if different app name
		if pkg.Spec.ArgoCD.Name != resource.Spec.ArgoCD.Name {
			continue
		}

		// Get the other package's priority
		otherPriority, err := getPackagePriority(pkg)
		if err != nil {
			// If the other package has no priority, this one wins
			continue
		}

		// If another package has higher priority, this one should not reconcile
		if otherPriority > thisPriority {
			logger.Info("yielding to higher priority package",
				"thisPackage", resource.Name,
				"thisPriority", thisPriority,
				"otherPackage", pkg.Name,
				"otherPriority", otherPriority,
				"appName", resource.Spec.ArgoCD.Name)
			return false, nil
		}
	}

	return true, nil
}

// getPackagePriority extracts the priority from a CustomPackage's annotations
func getPackagePriority(pkg *v1alpha1.CustomPackage) (int, error) {
	if pkg.ObjectMeta.Annotations == nil {
		return 0, fmt.Errorf("no annotations")
	}

	priorityStr, ok := pkg.ObjectMeta.Annotations[v1alpha1.PackagePriorityAnnotation]
	if !ok {
		return 0, fmt.Errorf("no priority annotation")
	}

	var priority int
	_, err := fmt.Sscanf(priorityStr, "%d", &priority)
	if err != nil {
		return 0, fmt.Errorf("invalid priority format: %w", err)
	}

	return priority, nil
}

// create an in-cluster repository CR, update the application spec, then apply
func (r *Reconciler) reconcileCustomPackage(ctx context.Context, resource *v1alpha1.CustomPackage) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Check if this package should reconcile
	shouldReconcile, err := r.shouldReconcile(ctx, resource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking if should reconcile: %w", err)
	}

	if !shouldReconcile {
		// This package has been superseded by a higher priority package
		// Mark as not synced and don't update resources, and don't requeue
		logger.Info("package superseded by higher priority, skipping reconciliation",
			"name", resource.Name,
			"appName", resource.Spec.ArgoCD.Name,
			"sourcePath", resource.ObjectMeta.Annotations[v1alpha1.PackageSourcePathAnnotation])
		resource.Status.Synced = false
		return ctrl.Result{}, nil
	}

	logger.V(1).Info("proceeding with reconciliation as highest priority package",
		"name", resource.Name,
		"appName", resource.Spec.ArgoCD.Name)

	b, err := r.getArgoCDAppFile(ctx, resource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reading file %s: %w", resource.Spec.ArgoCD.ApplicationFile, err)
	}

	objs, err := k8s.ConvertYamlToObjects(r.Scheme, b)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("converting yaml to object %w", err)
	}
	if len(objs) == 0 {
		return ctrl.Result{}, fmt.Errorf("file contained 0 kubernetes objects %s", resource.Spec.ArgoCD.ApplicationFile)
	}

	switch resource.Spec.ArgoCD.Type {
	case argocdapplication.ApplicationKind:
		app, ok := objs[0].(*argov1alpha1.Application)
		if !ok {
			return ctrl.Result{}, fmt.Errorf("object is not an ArgoCD application %s", resource.Spec.ArgoCD.ApplicationFile)
		}
		util.SetPackageLabels(app)

		res, err := r.reconcileArgoCDApp(ctx, resource, app)
		if err != nil {
			return ctrl.Result{}, err
		}

		foundAppObj := argov1alpha1.Application{}
		err = r.Client.Get(ctx, client.ObjectKeyFromObject(app), &foundAppObj)
		if err != nil {
			if errors.IsNotFound(err) {
				err = r.Client.Create(ctx, app)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("creating %s app CR: %w", app.Name, err)
				}

				return ctrl.Result{RequeueAfter: requeueTime}, nil
			}
			return ctrl.Result{}, fmt.Errorf("getting argocd application object: %w", err)
		}
		util.SetPackageLabels(&foundAppObj)
		foundAppObj.Spec = app.Spec
		foundAppObj.ObjectMeta.Annotations = app.GetAnnotations()
		foundAppObj.ObjectMeta.Labels = app.GetLabels()
		err = r.Client.Update(ctx, &foundAppObj)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("updating argocd application object %s: %w", app.Name, err)
		}
		return res, nil

	case argocdapplication.ApplicationSetKind:
		// application set embeds application spec. extract it then handle git generator repoURLs.
		appSet, ok := objs[0].(*argov1alpha1.ApplicationSet)
		if !ok {
			return ctrl.Result{}, fmt.Errorf("object is not an ArgoCD application set %s", resource.Spec.ArgoCD.ApplicationFile)
		}

		util.SetPackageLabels(appSet)

		res, err := r.reconcileArgoCDAppSet(ctx, resource, appSet)
		if err != nil {
			return ctrl.Result{}, err
		}

		foundAppSetObj := argov1alpha1.ApplicationSet{}
		err = r.Client.Get(ctx, client.ObjectKeyFromObject(appSet), &foundAppSetObj)
		if err != nil {
			if errors.IsNotFound(err) {
				err = r.Client.Create(ctx, appSet)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("creating %s argocd application set CR: %w", appSet.Name, err)
				}
				return ctrl.Result{RequeueAfter: requeueTime}, nil
			}
			return ctrl.Result{}, fmt.Errorf("getting argocd application set object: %w", err)
		}

		util.SetPackageLabels(&foundAppSetObj)
		foundAppSetObj.Spec = appSet.Spec
		foundAppSetObj.ObjectMeta.Annotations = appSet.GetAnnotations()
		foundAppSetObj.ObjectMeta.Labels = appSet.GetLabels()
		err = r.Client.Update(ctx, &foundAppSetObj)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("updating argocd application object %s: %w", appSet.Name, err)
		}
		return res, nil

	default:
		return ctrl.Result{}, fmt.Errorf("file is not a supported argocd kind %s", resource.Spec.ArgoCD.ApplicationFile)
	}
}

func (r *Reconciler) reconcileArgoCDApp(ctx context.Context, resource *v1alpha1.CustomPackage, app *argov1alpha1.Application) (ctrl.Result, error) {
	appSourcesSynced := true
	repoRefs := make([]v1alpha1.ObjectRef, 0, 1)
	if app.Spec.HasMultipleSources() {
		notSyncedRepos := 0
		for j := range app.Spec.Sources {
			s := &app.Spec.Sources[j]
			res, sErr := r.reconcileHelmValueObject(ctx, s, resource, app.Name)
			if sErr != nil {
				return res, sErr
			}

			res, repo, sErr := r.reconcileArgoCDSource(ctx, resource, s.RepoURL, app.Name)
			if sErr != nil {
				return res, sErr
			}

			if repo != nil {
				if repo.Status.InternalGitRepositoryUrl == "" {
					notSyncedRepos += 1
				}
				s.RepoURL = repo.Status.InternalGitRepositoryUrl
				repoRefs = append(repoRefs, v1alpha1.ObjectRef{
					Namespace: repo.Namespace,
					Name:      repo.Name,
					UID:       string(repo.ObjectMeta.UID),
				})
			}
		}
		appSourcesSynced = notSyncedRepos == 0
	} else {
		s := app.Spec.Source
		res, sErr := r.reconcileHelmValueObject(ctx, s, resource, app.Name)
		if sErr != nil {
			return res, sErr
		}

		res, repo, sErr := r.reconcileArgoCDSource(ctx, resource, s.RepoURL, app.Name)
		if sErr != nil {
			return res, sErr
		}

		if repo != nil {
			appSourcesSynced = repo.Status.InternalGitRepositoryUrl != ""
			s.RepoURL = repo.Status.InternalGitRepositoryUrl
			repoRefs = append(repoRefs, v1alpha1.ObjectRef{
				Namespace: repo.Namespace,
				Name:      repo.Name,
				UID:       string(repo.ObjectMeta.UID),
			})
		}
	}
	resource.Status.GitRepositoryRefs = repoRefs
	resource.Status.Synced = appSourcesSynced

	// Only requeue if not synced yet to avoid continuous reconciliation
	if !appSourcesSynced {
		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileArgoCDAppSet(ctx context.Context, resource *v1alpha1.CustomPackage, appSet *argov1alpha1.ApplicationSet) (ctrl.Result, error) {
	notSyncedRepos := 0
	for i := range appSet.Spec.Generators {
		g := appSet.Spec.Generators[i]
		if g.Git != nil {
			res, repo, gErr := r.reconcileArgoCDSource(ctx, resource, g.Git.RepoURL, appSet.GetName())
			if gErr != nil {
				return res, fmt.Errorf("reconciling git generator URL %s, %s: %w", g.Git.RepoURL, resource.Spec.ArgoCD.ApplicationFile, gErr)
			}
			if repo != nil {
				g.Git.RepoURL = repo.Status.InternalGitRepositoryUrl
				if repo.Status.InternalGitRepositoryUrl == "" {
					notSyncedRepos += 1
				}
			}
		}
		if g.Matrix != nil {
			for j := range g.Matrix.Generators {
				nestedGenerator := g.Matrix.Generators[j]
				if nestedGenerator.Git != nil {
					res, repo, gErr := r.reconcileArgoCDSource(ctx, resource, nestedGenerator.Git.RepoURL, appSet.GetName())
					if gErr != nil {
						return res, fmt.Errorf("reconciling git generator URL %s, %s: %w", nestedGenerator.Git.RepoURL, resource.Spec.ArgoCD.ApplicationFile, gErr)
					}
					if repo != nil {
						nestedGenerator.Git.RepoURL = repo.Status.InternalGitRepositoryUrl
						if repo.Status.InternalGitRepositoryUrl == "" {
							notSyncedRepos += 1
						}
					}
				}
			}
		}
	}

	gitGeneratorsSynced := notSyncedRepos == 0
	app := argov1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: appSet.GetName(), Namespace: appSet.Namespace},
	}
	app.Spec = appSet.Spec.Template.Spec

	_, err := r.reconcileArgoCDApp(ctx, resource, &app)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling application set %s %w", resource.Spec.ArgoCD.ApplicationFile, err)
	}

	resource.Status.Synced = resource.Status.Synced && gitGeneratorsSynced

	// Only requeue if not synced yet to avoid continuous reconciliation
	if !resource.Status.Synced {
		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}
	return ctrl.Result{}, nil
}

// create a gitrepository custom resource, then let the git repository controller take care of the rest
func (r *Reconciler) reconcileArgoCDSource(ctx context.Context, resource *v1alpha1.CustomPackage, repoUrl, appName string) (ctrl.Result, *v1alpha1.GitRepository, error) {
	if isCNOEScheme(repoUrl) {
		if resource.Spec.RemoteRepository.Url == "" {
			return r.reconcileArgoCDSourceFromLocal(ctx, resource, appName, repoUrl)
		}
		return r.reconcileArgoCDSourceFromRemote(ctx, resource, appName, repoUrl)
	}
	return ctrl.Result{}, nil, nil
}

func (r *Reconciler) reconcileArgoCDSourceFromRemote(ctx context.Context, resource *v1alpha1.CustomPackage, appName, repoURL string) (ctrl.Result, *v1alpha1.GitRepository, error) {
	logger := log.FromContext(ctx)
	relativePath := strings.TrimPrefix(repoURL, v1alpha1.CNOEURIScheme)
	// no guarantee that this path exists
	dirPath := filepath.Join(resource.Spec.RemoteRepository.Path, relativePath)

	repo := &v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      remoteRepoName(appName, dirPath, resource.Spec.RemoteRepository),
			Namespace: resource.Namespace,
		},
	}

	// Check if GitRepository already exists
	existingRepo := &v1alpha1.GitRepository{}
	getErr := r.Client.Get(ctx, client.ObjectKeyFromObject(repo), existingRepo)

	if getErr == nil {
		// GitRepository exists - check if we should take it over
		shouldTakeOver, checkErr := r.shouldTakeOverGitRepository(ctx, resource, existingRepo)
		if checkErr != nil {
			return ctrl.Result{}, nil, fmt.Errorf("checking if should take over git repository: %w", checkErr)
		}

		if !shouldTakeOver {
			// A higher/equal priority package owns this, just return it without updating
			logger.V(1).Info("Using existing GitRepository owned by higher priority package",
				"repo", repo.Name,
				"ourPriority", resource.ObjectMeta.Annotations[v1alpha1.PackagePriorityAnnotation])
			return ctrl.Result{}, existingRepo, nil
		}
		// We should take it over - proceed with CreateOrUpdate below
	} else if !errors.IsNotFound(getErr) {
		return ctrl.Result{}, nil, getErr
	}
	// GitRepository doesn't exist or we should take it over

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
				Type:             v1alpha1.SourceTypeRemote,
				RemoteRepository: resource.Spec.RemoteRepository,
				Path:             dirPath,
			},
			Provider: v1alpha1.Provider{
				Name:             v1alpha1.GitProviderGitea,
				GitURL:           resource.Spec.GitServerURL,
				InternalGitURL:   resource.Spec.InternalGitServeURL,
				OrganizationName: v1alpha1.GiteaAdminUserName,
			},
			SecretRef: resource.Spec.GitServerAuthSecretRef,
		}

		return nil
	})

	if err != nil && !errors.IsAlreadyExists(err) {
		return ctrl.Result{}, nil, err
	}

	return ctrl.Result{}, repo, nil
}

func (r *Reconciler) reconcileArgoCDSourceFromLocal(ctx context.Context, resource *v1alpha1.CustomPackage, appName, repoURL string) (ctrl.Result, *v1alpha1.GitRepository, error) {
	logger := log.FromContext(ctx)

	absPath, err := getCNOEAbsPath(resource.Spec.ArgoCD.ApplicationFile, repoURL)
	if err != nil {
		logger.Error(err, "processing argocd app source", "dir", resource.Spec.ArgoCD.ApplicationFile, "repoURL", repoURL)
		return ctrl.Result{}, nil, err
	}

	repo := &v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      localRepoName(appName, absPath),
			Namespace: resource.Namespace,
		},
	}

	// Check if GitRepository already exists
	existingRepo := &v1alpha1.GitRepository{}
	getErr := r.Client.Get(ctx, client.ObjectKeyFromObject(repo), existingRepo)

	if getErr == nil {
		// GitRepository exists - check if we should take it over
		shouldTakeOver, checkErr := r.shouldTakeOverGitRepository(ctx, resource, existingRepo)
		if checkErr != nil {
			return ctrl.Result{}, nil, fmt.Errorf("checking if should take over git repository: %w", checkErr)
		}

		if !shouldTakeOver {
			// A higher/equal priority package owns this, just return it without updating
			logger.V(1).Info("Using existing GitRepository owned by higher priority package",
				"repo", repo.Name,
				"ourPriority", resource.ObjectMeta.Annotations[v1alpha1.PackagePriorityAnnotation])
			return ctrl.Result{}, existingRepo, nil
		}
		// We should take it over - proceed with CreateOrUpdate below
	} else if !errors.IsNotFound(getErr) {
		return ctrl.Result{}, nil, getErr
	}
	// GitRepository doesn't exist or we should take it over

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
				Type: v1alpha1.SourceTypeLocal,
				Path: absPath,
			},
			Provider: v1alpha1.Provider{
				Name:             v1alpha1.GitProviderGitea,
				GitURL:           resource.Spec.GitServerURL,
				InternalGitURL:   resource.Spec.InternalGitServeURL,
				OrganizationName: v1alpha1.GiteaAdminUserName,
			},
			SecretRef: resource.Spec.GitServerAuthSecretRef,
		}

		return nil
	})
	// it's possible for an application to specify the same directory multiple times in the spec.
	// if there is a repository already created for this package, no further action is necessary.
	if !errors.IsAlreadyExists(err) {
		return ctrl.Result{}, repo, err
	}

	return ctrl.Result{}, repo, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CustomPackage{}).
		Complete(r)
}

func (r *Reconciler) getArgoCDAppFile(ctx context.Context, resource *v1alpha1.CustomPackage) ([]byte, error) {
	filePath := resource.Spec.ArgoCD.ApplicationFile

	if resource.Spec.RemoteRepository.Url == "" {
		return os.ReadFile(filePath)
	}

	cloneDir := util.RepoDir(resource.Spec.RemoteRepository.Url, r.TempDir)
	st := r.RepoMap.LoadOrStore(resource.Spec.RemoteRepository.Url, cloneDir)
	st.MU.Lock()
	wt, _, err := util.CloneRemoteRepoToDir(ctx, resource.Spec.RemoteRepository, 1, false, cloneDir, "")
	defer st.MU.Unlock()
	if err != nil {
		return nil, fmt.Errorf("cloning repo, %s: %w", resource.Spec.RemoteRepository.Url, err)
	}
	return util.ReadWorktreeFile(wt, filePath)
}

func (r *Reconciler) reconcileHelmValueObject(ctx context.Context, source *argov1alpha1.ApplicationSource,
	resource *v1alpha1.CustomPackage, appName string,
) (ctrl.Result, error) {
	if source.Helm == nil || source.Helm.ValuesObject == nil {
		return ctrl.Result{}, nil
	}

	var data any
	err := json.Unmarshal(source.Helm.ValuesObject.Raw, &data)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("processing helm valuesObject: %w", err)
	}

	res, err := r.reconcileHelmValueObjectSource(ctx, &data, resource, appName)
	if err != nil {
		return ctrl.Result{}, err
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("converting helm valuesObject to json")
	}

	source.Helm.ValuesObject.Raw = raw
	return res, nil
}

func (r *Reconciler) reconcileHelmValueObjectSource(ctx context.Context,
	valueObject *any, resource *v1alpha1.CustomPackage, appName string,
) (ctrl.Result, error) {

	switch val := (*valueObject).(type) {
	case string:
		res, repo, err := r.reconcileArgoCDSource(ctx, resource, val, appName)
		if err != nil {
			return res, fmt.Errorf("processing %s in helmValueObject: %w", val, err)
		}
		if repo != nil {
			*valueObject = repo.Status.InternalGitRepositoryUrl
		}
	case map[string]any:
		for k := range val {
			v := val[k]
			res, err := r.reconcileHelmValueObjectSource(ctx, &v, resource, appName)
			if err != nil {
				return res, err
			}
			val[k] = v
		}
	case []any:
		for k := range val {
			v := val[k]
			res, err := r.reconcileHelmValueObjectSource(ctx, &v, resource, appName)
			if err != nil {
				return res, err
			}
			val[k] = v
		}
	}
	// No need to requeue for helm value processing
	return ctrl.Result{}, nil
}

func localRepoName(appName, dir string) string {
	return fmt.Sprintf("%s-%s", appName, filepath.Base(dir))
}

func remoteRepoName(appName, pathToPkg string, repo v1alpha1.RemoteRepositorySpec) string {
	return fmt.Sprintf("%s-%s", appName, filepath.Base(pathToPkg))
}

func isCNOEScheme(repoURL string) bool {
	return strings.HasPrefix(repoURL, v1alpha1.CNOEURIScheme)
}

func getCNOEAbsPath(fPath, repoURL string) (string, error) {
	parentDir := filepath.Dir(fPath)
	relativePath := strings.TrimPrefix(repoURL, v1alpha1.CNOEURIScheme)
	absPath, err := filepath.Abs(filepath.Join(parentDir, relativePath))
	if err != nil {
		return "", err
	}

	f, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}
	if !f.IsDir() {
		return "", fmt.Errorf("path not a directory: %s", absPath)
	}
	return absPath, err
}
