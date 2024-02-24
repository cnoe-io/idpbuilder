package localbuild

import (
	"context"
	"embed"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// hardcoded values from what we have in the yaml installation file.
	giteaNamespace   = "gitea"
	giteaAdminSecret = "gitea-admin-secret"
	// this is the URL accessible outside cluster. resolves to localhost
	giteaIngressURL = "https://gitea.cnoe.localtest.me:%s"
	// this is the URL accessible within cluster for ArgoCD to fetch resources.
	// resolves to cluster ip
	giteaSvcURL = "http://my-gitea-http.gitea.svc.cluster.local:3000"
)

//go:embed resources/gitea/k8s/*
var installGiteaFS embed.FS

func RawGiteaInstallResources(tmpl interface{}) ([][]byte, error) {
	return util.ConvertFSToBytes(installGiteaFS, "resources/gitea/k8s", tmpl)
}

func (r *LocalbuildReconciler) ReconcileGitea(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
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

	if result, err := gitea.Install(ctx, req, resource, r.Client, r.Scheme, r.Config); err != nil {
		return result, err
	}
	resource.Status.Gitea.ExternalURL = fmt.Sprintf(giteaIngressURL, r.Config.Port)
	resource.Status.Gitea.InternalURL = giteaSvcURL
	resource.Status.Gitea.AdminUserSecretName = giteaAdminSecret
	resource.Status.Gitea.AdminUserSecretNamespace = giteaNamespace
	resource.Status.Gitea.Available = true
	return ctrl.Result{}, nil
}
