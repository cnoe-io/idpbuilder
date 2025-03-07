package localbuild

import (
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"net/http"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed resources/gitea/k8s/*
var installGiteaFS embed.FS

func RawGiteaInstallResources(templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	return k8s.BuildCustomizedManifests(config.FilePath, "resources/gitea/k8s", installGiteaFS, scheme, templateData)
}

func (r *LocalbuildReconciler) newGiteaAdminSecret(password string) corev1.Secret {
	obj := util.GiteaAdminSecretObject()
	obj.StringData = map[string]string{
		"username": v1alpha1.GiteaAdminUserName,
		"password": password,
	}
	return obj
}

func (r *LocalbuildReconciler) ReconcileGitea(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "installer", "gitea")
	gitea := EmbeddedInstallation{
		name:         "Gitea",
		resourcePath: "resources/gitea/k8s",
		resourceFS:   installGiteaFS,
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
			gitea.unmanagedResources = []client.Object{&giteaCreds}
			sec = giteaCreds
		} else {
			return ctrl.Result{}, fmt.Errorf("getting gitea secret: %w", err)
		}
	}

	v, ok := resource.Spec.PackageConfigs.CorePackageCustomization[v1alpha1.GiteaPackageName]
	if ok {
		gitea.customization = v
	}

	if result, err := gitea.Install(ctx, resource, r.Client, r.Scheme, r.Config); err != nil {
		return result, err
	}

	baseUrl, err := util.GiteaBaseUrl(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	// need this to ensure gitrepository controller can reach the api endpoint.
	logger.V(1).Info("checking gitea api endpoint", "url", baseUrl)
	c := util.GetHttpClient()
	resp, err := c.Get(baseUrl)
	if err != nil {
		return ctrl.Result{}, err
	}
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.V(1).Info("gitea manifests installed successfully. endpoint not ready", "statusCode", resp.StatusCode)
			return ctrl.Result{RequeueAfter: errRequeueTime}, nil
		}
	}

	err = r.setGiteaToken(ctx, sec, baseUrl)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating gitea token: %w", err)
	}

	resource.Status.Gitea.ExternalURL = baseUrl
	resource.Status.Gitea.InternalURL = giteaInternalBaseUrl(r.Config)
	resource.Status.Gitea.AdminUserSecretName = util.GiteaAdminSecret
	resource.Status.Gitea.AdminUserSecretNamespace = util.GiteaNamespace
	resource.Status.Gitea.Available = true
	return ctrl.Result{}, nil
}

func (r *LocalbuildReconciler) setGiteaToken(ctx context.Context, secret corev1.Secret, baseUrl string) error {
	_, ok := secret.Data[util.GiteaAdminTokenFieldName]
	if ok {
		return nil
	}

	u := unstructured.Unstructured{}
	u.SetName(util.GiteaAdminSecret)
	u.SetNamespace(util.GiteaNamespace)
	u.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

	user, ok := secret.Data["username"]
	if !ok {
		return fmt.Errorf("username field not found in gitea secret")
	}

	pass, ok := secret.Data["password"]
	if !ok {
		return fmt.Errorf("password field not found in gitea secret")
	}

	t, err := util.GetGiteaToken(ctx, baseUrl, string(user), string(pass))
	if err != nil {
		return fmt.Errorf("getting gitea token: %w", err)
	}

	token := base64.StdEncoding.EncodeToString([]byte(t))
	err = unstructured.SetNestedField(u.Object, token, "data", util.GiteaAdminTokenFieldName)
	if err != nil {
		return fmt.Errorf("setting gitea token field: %w", err)
	}

	return r.Client.Patch(ctx, &u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}

// gitea URL reachable within the cluster with proper coredns config. Mainly for argocd
func giteaInternalBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	if config.UsePathRouting {
		return fmt.Sprintf(util.GiteaURLTempl, config.Protocol, "", config.Host, config.Port, "/gitea")
	}
	return fmt.Sprintf(util.GiteaURLTempl, config.Protocol, "gitea.", config.Host, config.Port, "")
}
