package localbuild

import (
	"context"
	"fmt"
	"time"

	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/apps"
	"github.com/cnoe-io/idpbuilder/pkg/resources/localbuild"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultArgoCDProjectName     string = "default"
	EmbeddedGitServerName        string = "embedded"
	gitServerIngressHostnameBase string = ".cnoe.localtest.me"
	repoUrlFmt                   string = "http://%s.%s.svc/idpbuilder-resources.git"
)

func getRepoUrl(resource *v1alpha1.GitServer) string {
	return fmt.Sprintf(repoUrlFmt, managedResourceName(resource), resource.Namespace)
}

var gitServerLabelKey string = fmt.Sprintf("%s-gitserver", globals.ProjectName)

func ingressHostname(resource *v1alpha1.GitServer) string {
	return fmt.Sprintf("%s%s", resource.Name, gitServerIngressHostnameBase)
}

func managedResourceName(resource *v1alpha1.GitServer) string {
	return fmt.Sprintf("%s-%s", globals.GitServerResourcename(), resource.Name)
}

type LocalbuildReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	CancelFunc     context.CancelFunc
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
	}

	switch localBuild.Spec.PackageConfigs.GitConfig.Type {
	case globals.GitServerResourcename():
		subReconcilers = append(
			subReconcilers,
			[]subReconciler{r.ReconcileEmbeddedGitServer, r.ReconcileArgoAppsWithGitServer}...,
		)
	case globals.GiteaResourceName():
		subReconcilers = append(
			subReconcilers,
			[]subReconciler{r.ReconcileGitea, r.ReconcileArgoAppsWithGitea}...,
		)
	default:
		return ctrl.Result{}, fmt.Errorf("GitConfig %s is invalid for LocalBuild %s", localBuild.Spec.PackageConfigs.GitConfig.Type, localBuild.GetName())
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

func (r *LocalbuildReconciler) ReconcileEmbeddedGitServer(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Bail if argo is not yet available
	if !resource.Status.ArgoAvailable {
		log.Info("argo not yet available, not installing embedded git server")
		return ctrl.Result{}, nil
	}

	// Bail if embedded argo applications not enabled
	if !resource.Spec.PackageConfigs.EmbeddedArgoApplications.Enabled {
		log.Info("embedded argo applications disabled, not installing embedded git server")
		return ctrl.Result{}, nil
	}

	gitServerResource := &v1alpha1.GitServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EmbeddedGitServerName,
			Namespace: globals.GetProjectNamespace(resource.Name),
		},
	}

	log.Info("Create or update git server", "resource", gitServerResource)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, gitServerResource, func() error {
		if err := controllerutil.SetControllerReference(resource, gitServerResource, r.Scheme); err != nil {
			log.Error(err, "Setting controller ref on git server resource")
			return err
		}

		gitServerResource.Spec.Source.Embedded = true
		return nil
	})
	if err != nil {
		log.Error(err, "Create or Update git server resource")
	}

	// Bail if the GitServer deployment is not yet available
	if !gitServerResource.Status.DeploymentAvailable {
		log.Info("Waiting for GitServer to become available before creating argo applications")
		return ctrl.Result{
			RequeueAfter: time.Second * 10,
		}, nil
	}

	return ctrl.Result{}, err
}

func (r *LocalbuildReconciler) ReconcileArgoAppsWithGitServer(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Bail if embedded argo applications not enabled
	if !resource.Spec.PackageConfigs.EmbeddedArgoApplications.Enabled {
		log.Info("embedded argo applications disabled, not installing embedded git server")
		r.shouldShutdown = true
		return ctrl.Result{}, nil
	}

	// Create argo project
	// DeepEqual is broken on argo resources for some reason so we have to DIY create/update
	project := &argov1alpha1.AppProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.GetArgoProjectName(),
			Namespace: "argocd",
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(project), project); err != nil {
		localbuild.SetProjectSpec(project)
		log.Info("Creating project", "resource", project)
		if err := r.Client.Create(ctx, project); err != nil {
			log.Error(err, "Creating argo project", "resource", project)
			return ctrl.Result{}, err
		}
	}

	foundGitServer := &v1alpha1.GitServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EmbeddedGitServerName,
			Namespace: globals.GetProjectNamespace(resource.Name),
		},
	}

	err := r.Client.Get(ctx, client.ObjectKeyFromObject(foundGitServer), foundGitServer)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Could not find GitServer")
		return ctrl.Result{}, err
	}

	if !foundGitServer.Spec.Source.Embedded {
		log.Info("Not using embedded source, skipping argo app creation")
		return ctrl.Result{}, nil
	}

	// Install Argo Apps
	for _, embedApp := range apps.EmbedApps {
		log.Info("Ensuring Argo Application", "name", embedApp.Name)
		app := &argov1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resource.GetArgoApplicationName(embedApp.Name),
				Namespace: "argocd",
			},
		}

		if err := controllerutil.SetControllerReference(resource, app, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Client.Get(ctx, client.ObjectKeyFromObject(app), app); err != nil {
			log.Info("Argo app doesnt exist, creating", "name", embedApp.Name)
			repoUrl := getRepoUrl(foundGitServer)

			localbuild.SetApplicationSpec(
				app,
				repoUrl,
				embedApp.Path,
				defaultArgoCDProjectName,
				"argocd",
				nil,
			)

			if err := r.Client.Create(ctx, app); err != nil {
				log.Error(err, "Creating argo app", "resource", app)
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Argo app exists, skipping", "name", embedApp.Name)
		}
	}

	resource.Status.ArgoAppsCreated = true
	r.shouldShutdown = true

	return ctrl.Result{}, nil
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
	// do the same for embedded applications
	for _, embedApp := range apps.EmbedApps {
		result, err := r.reconcileEmbeddedApp(ctx, embedApp.Name, resource)
		if err != nil {
			return result, fmt.Errorf("reconciling embedded apps %w", err)
		}
	}

	// TODO: this needs to be removed for local file syncs.
	r.shouldShutdown = true
	return ctrl.Result{}, nil
}

func (r *LocalbuildReconciler) reconcileEmbeddedApp(ctx context.Context, appName string, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Ensuring Argo Application", "name", appName)
	repo := &v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: globals.GetProjectNamespace(resource.Name),
		},
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				EmbeddedAppName: appName,
				Type:            "embedded",
			},
			GitURL: resource.Status.Gitea.ExternalURL,
			SecretRef: v1alpha1.SecretReference{
				Name:      resource.Status.Gitea.AdminUserSecretName,
				Namespace: resource.Status.Gitea.AdminUserSecretNamespace,
			},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, repo, func() error {
		if err := controllerutil.SetControllerReference(resource, repo, r.Scheme); err != nil {
			return err
		}
		return nil
	})
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
			getRepositoryURL(repo.Namespace, repo.Name, resource.Status.Gitea.InternalURL),
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

	return ctrl.Result{}, nil
}

func GetEmbeddedRawInstallResources(name string) ([][]byte, error) {
	switch name {
	case "argocd":
		return RawArgocdInstallResources()
	case "backstage", "crossplane":
		return util.ConvertFSToBytes(apps.EmbeddedAppsFS, fmt.Sprintf("srv/%s", name))
	case "gitea":
		return RawGiteaInstallResources()
	case "nginx":
		return RawNginxInstallResources()
	default:
		return nil, fmt.Errorf("unsupported embedded app name %s", name)
	}
}
