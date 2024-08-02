package localbuild

import (
	"context"
	"embed"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

//go:embed resources/argo/*
var installArgoFS embed.FS

func RawArgocdInstallResources(templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	return k8s.BuildCustomizedManifests(config.FilePath, "resources/argo", installArgoFS, scheme, templateData)
}

func (r *LocalbuildReconciler) ReconcileArgo(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	argocd := EmbeddedInstallation{
		name:         "Argo CD",
		resourcePath: "resources/argo",
		resourceFS:   installArgoFS,
		namespace:    globals.ArgoCDNamespace,
		monitoredResources: map[string]schema.GroupVersionKind{
			"argocd-server": {
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			"argocd-repo-server": {
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			"argocd-application-controller": {
				Group:   "apps",
				Version: "v1",
				Kind:    "StatefulSet",
			},
		},
		skipReadinessCheck: true,
	}

	v, ok := resource.Spec.PackageConfigs.CorePackageCustomization[v1alpha1.ArgoCDPackageName]
	if ok {
		argocd.customization = v
	}

	if result, err := argocd.Install(ctx, resource, r.Client, r.Scheme, r.Config); err != nil {
		return result, err
	}

	resource.Status.ArgoCD.Available = true
	return ctrl.Result{}, nil
}
