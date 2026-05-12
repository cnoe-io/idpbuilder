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

//go:embed resources/traefik/k8s/*
var installTraefikFS embed.FS

func RawTraefikInstallResources(templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	return k8s.BuildCustomizedManifests(config.FilePath, "resources/traefik/k8s", installTraefikFS, scheme, templateData)
}

func (r *LocalbuildReconciler) ReconcileTraefik(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	traefik := EmbeddedInstallation{
		name:         "Traefik",
		resourcePath: "resources/traefik/k8s",
		resourceFS:   installTraefikFS,
		namespace:    globals.TraefikNamespace,
		monitoredResources: map[string]schema.GroupVersionKind{
			"traefik": {
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
		},
	}

	v, ok := resource.Spec.PackageConfigs.CorePackageCustomization[v1alpha1.TraefikPackageName]
	if ok {
		traefik.customization = v
	}

	if result, err := traefik.Install(ctx, resource, r.Client, r.Scheme, r.Config); err != nil {
		return result, err
	}

	resource.Status.Traefik.Available = true
	return ctrl.Result{}, nil
}
