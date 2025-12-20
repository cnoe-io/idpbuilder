package localbuild

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/resources/gitea"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func RawGiteaInstallResources(templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	return gitea.RawGiteaInstallResources(templateData, config, scheme)
}

func (r *LocalbuildReconciler) newGiteaAdminSecret(password string) corev1.Secret {
	return gitea.NewGiteaAdminSecret(password)
}

func (r *LocalbuildReconciler) ReconcileGitea(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "installer", "gitea")
	giteaInstall := EmbeddedInstallation{
		name:         "Gitea",
		resourcePath: "resources/k8s",
		resourceFS:   gitea.GetInstallFS(),
		namespace:    util.GiteaNamespace,
		monitoredResources: map[string]schema.GroupVersionKind{
			"my-gitea": {
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
		},
	}

	sec := util.GiteaAdminSecretObject()
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: sec.GetNamespace(),
		Name:      sec.GetName(),
	}, &sec)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			genPassword, err := util.GeneratePassword()
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("generating gitea password: %w", err)
			}

			giteaCreds := r.newGiteaAdminSecret(genPassword)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("generating gitea admin secret: %w", err)
			}
			giteaInstall.unmanagedResources = []client.Object{&giteaCreds}
			sec = giteaCreds
		} else {
			return ctrl.Result{}, fmt.Errorf("getting gitea secret: %w", err)
		}
	}

	v, ok := resource.Spec.PackageConfigs.CorePackageCustomization[v1alpha1.GiteaPackageName]
	if ok {
		giteaInstall.customization = v
	}

	if result, err := giteaInstall.Install(ctx, resource, r.Client, r.Scheme, r.Config); err != nil {
		return result, err
	}

	baseUrl := util.GiteaBaseUrl(r.Config)

	// need this to ensure gitrepository controller can reach the api endpoint.
	logger.V(1).Info("checking gitea api endpoint", "url", baseUrl)
	ready, err := gitea.CheckGiteaEndpoint(baseUrl)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !ready {
		logger.V(1).Info("gitea manifests installed successfully. endpoint not ready")
		return ctrl.Result{RequeueAfter: errRequeueTime}, nil
	}

	err = gitea.SetGiteaToken(ctx, r.Client, sec, baseUrl)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating gitea token: %w", err)
	}

	resource.Status.Gitea.ExternalURL = baseUrl
	resource.Status.Gitea.InternalURL = util.GiteaBaseUrl(r.Config)
	resource.Status.Gitea.AdminUserSecretName = util.GiteaAdminSecret
	resource.Status.Gitea.AdminUserSecretNamespace = util.GiteaNamespace
	resource.Status.Gitea.Available = true
	return ctrl.Result{}, nil
}
