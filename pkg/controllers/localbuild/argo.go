package localbuild

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

//go:embed resources/argo/*
var installArgoFS embed.FS

const (
	argocdDevModePassword = "developer"
)

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

	// Let's patch the existing argocd admin secret if devmode is enabled to set the default password
	if r.Config.DevMode {
		kubeClient, err := k8s.GetKubeClient()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting kube client: %w", err)
		}

		s := v1.Secret{}
		err = kubeClient.Get(ctx, client.ObjectKey{Name: "argocd-secret", Namespace: "argocd"}, &s)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting argocd secret: %w", err)
		}

		// Hash password using bcrypt
		password := argocdDevModePassword
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 0)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Error hashing password: %w", err)
		}
		// Get the current date in the desired format
		passwordMtime := time.Now().Format("2006-01-02T15:04:05Z")

		// Prepare the patch for the Secret's `stringData` field
		patchData := map[string]interface{}{
			"stringData": map[string]string{
				"admin.password":      string(hashedPassword),
				"admin.passwordMtime": passwordMtime,
			},
		}
		// Convert patch data to JSON
		patchBytes, err := json.Marshal(patchData)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Error marshalling patch data:", err)
		}

		// Patching the argocd-secret with the hashed password
		err = kubeClient.Patch(ctx, &s, client.RawPatch(types.StrategicMergePatchType, patchBytes))
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Error patching the Secret:", err)
		} else {
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}
