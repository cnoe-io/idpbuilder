package localbuild

import (
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"net/http"

	"code.gitea.io/sdk/gitea"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	giteaDevModePassword = "developer"

	// hardcoded values from what we have in the yaml installation file.
	giteaNamespace           = "gitea"
	giteaAdminSecret         = "gitea-credential"
	giteaAdminTokenName      = "admin"
	giteaAdminTokenFieldName = "token"
	// this is the URL accessible outside cluster. resolves to localhost
	giteaIngressURL = "%s://gitea.cnoe.localtest.me:%s"
	// this is the URL accessible within cluster for ArgoCD to fetch resources.
	// resolves to cluster ip
	giteaSvcURL = "%s://%s%s:%s%s"
)

//go:embed resources/gitea/k8s/*
var installGiteaFS embed.FS

func RawGiteaInstallResources(templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	return k8s.BuildCustomizedManifests(config.FilePath, "resources/gitea/k8s", installGiteaFS, scheme, templateData)
}

func giteaAdminSecretObject() corev1.Secret {
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      giteaAdminSecret,
			Namespace: giteaNamespace,
		},
	}
}

func newGiteaAdminSecret(devMode bool) (corev1.Secret, error) {
	pass := giteaDevModePassword
	// TODO: Reverting to giteaAdmin till we know why a different user - developer fails
	userName := v1alpha1.GiteaAdminUserName

	if !devMode {
		var err error
		pass, err = util.GeneratePassword()
		if err != nil {
			return corev1.Secret{}, err
		}

		userName = v1alpha1.GiteaAdminUserName
	}

	obj := giteaAdminSecretObject()
	obj.StringData = map[string]string{
		"username": userName,
		"password": pass,
	}
	return obj, nil
}

func (r *LocalbuildReconciler) ReconcileGitea(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "installer", "gitea")
	gitea := EmbeddedInstallation{
		name:         "Gitea",
		resourcePath: "resources/gitea/k8s",
		resourceFS:   installGiteaFS,
		namespace:    giteaNamespace,
		monitoredResources: map[string]schema.GroupVersionKind{
			"my-gitea": {
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
		},
	}

	sec := giteaAdminSecretObject()
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: sec.GetNamespace(),
		Name:      sec.GetName(),
	}, &sec)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			giteaCreds, err := newGiteaAdminSecret(r.Config.DevMode)
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

	baseUrl := giteaBaseUrl(r.Config)
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
	resource.Status.Gitea.AdminUserSecretName = giteaAdminSecret
	resource.Status.Gitea.AdminUserSecretNamespace = giteaNamespace
	resource.Status.Gitea.Available = true
	return ctrl.Result{}, nil
}

func (r *LocalbuildReconciler) setGiteaToken(ctx context.Context, secret corev1.Secret, baseUrl string) error {
	_, ok := secret.Data[giteaAdminTokenFieldName]
	if ok {
		return nil
	}

	u := unstructured.Unstructured{}
	u.SetName(giteaAdminSecret)
	u.SetNamespace(giteaNamespace)
	u.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

	user, ok := secret.Data["username"]
	if !ok {
		return fmt.Errorf("username field not found in gitea secret")
	}

	pass, ok := secret.Data["password"]
	if !ok {
		return fmt.Errorf("password field not found in gitea secret")
	}

	t, err := getGiteaToken(ctx, baseUrl, string(user), string(pass))
	if err != nil {
		return fmt.Errorf("getting gitea token: %w", err)
	}

	token := base64.StdEncoding.EncodeToString([]byte(t))
	err = unstructured.SetNestedField(u.Object, token, "data", giteaAdminTokenFieldName)
	if err != nil {
		return fmt.Errorf("setting gitea token field: %w", err)
	}

	return r.Client.Patch(ctx, &u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}

func getGiteaToken(ctx context.Context, baseUrl, username, password string) (string, error) {
	giteaClient, err := gitea.NewClient(baseUrl, gitea.SetHTTPClient(util.GetHttpClient()),
		gitea.SetBasicAuth(username, password), gitea.SetContext(ctx),
	)
	if err != nil {
		return "", fmt.Errorf("creating gitea client: %w", err)
	}
	tokens, resp, err := giteaClient.ListAccessTokens(gitea.ListAccessTokensOptions{})
	if err != nil {
		return "", fmt.Errorf("listing gitea access tokens. status: %s error : %w", resp.Status, err)
	}

	for i := range tokens {
		if tokens[i].Name == giteaAdminTokenName {
			resp, err := giteaClient.DeleteAccessToken(tokens[i].ID)
			if err != nil {
				return "", fmt.Errorf("deleting gitea access tokens. status: %s error : %w", resp.Status, err)
			}
			break
		}
	}

	token, resp, err := giteaClient.CreateAccessToken(gitea.CreateAccessTokenOption{
		Name: giteaAdminTokenName,
		Scopes: []gitea.AccessTokenScope{
			gitea.AccessTokenScopeAll,
		},
	})
	if err != nil {
		return "", fmt.Errorf("deleting gitea access tokens. status: %s error : %w", resp.Status, err)
	}

	return token.Token, nil
}

func giteaBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	return fmt.Sprintf(giteaIngressURL, config.Protocol, config.Port)
}

// gitea URL reachable within the cluster with proper coredns config. Mainly for argocd
func giteaInternalBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	if config.UsePathRouting {
		return fmt.Sprintf(giteaSvcURL, config.Protocol, "", config.Host, config.Port, "/gitea")
	}
	return fmt.Sprintf(giteaSvcURL, config.Protocol, "gitea.", config.Host, config.Port, "")
}
