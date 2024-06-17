package localbuild

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// hardcoded values from what we have in the yaml installation file.
	giteaNamespace   = "gitea"
	giteaAdminSecret = "gitea-credential"
	// this is the URL accessible outside cluster. resolves to localhost
	giteaIngressURL = "%s://gitea.cnoe.localtest.me:%s"
	// this is the URL accessible within cluster for ArgoCD to fetch resources.
	// resolves to cluster ip
	giteaSvcURL = "http://my-gitea-http.gitea.svc.cluster.local:3000"
)

//go:embed resources/gitea/k8s/*
var installGiteaFS embed.FS

func RawGiteaInstallResources(templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	return k8s.BuildCustomizedManifests(config.FilePath, "resources/gitea/k8s", installGiteaFS, scheme, templateData)
}

func newGiteaAdminSecret() (corev1.Secret, error) {
	pass, err := util.GeneratePassword()
	if err != nil {
		return corev1.Secret{}, err
	}
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      giteaAdminSecret,
			Namespace: giteaNamespace,
		},
		StringData: map[string]string{
			"username": v1alpha1.GiteaAdminUserName,
			"password": pass,
		},
	}, nil
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

	giteCreds, err := newGiteaAdminSecret()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("generating gitea admin secret: %w", err)
	}

	gitea.unmanagedResources = []client.Object{&giteCreds}

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
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 2 * time.Second,
	}
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

	resource.Status.Gitea.ExternalURL = baseUrl
	resource.Status.Gitea.InternalURL = giteaSvcURL
	resource.Status.Gitea.AdminUserSecretName = giteaAdminSecret
	resource.Status.Gitea.AdminUserSecretNamespace = giteaNamespace
	resource.Status.Gitea.Available = true
	return ctrl.Result{}, nil
}

func giteaBaseUrl(config util.CorePackageTemplateConfig) string {
	return fmt.Sprintf(giteaIngressURL, config.Protocol, config.Port)
}
