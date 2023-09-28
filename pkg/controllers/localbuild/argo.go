package localbuild

import (
	"context"
	"embed"

	"git.autodesk.com/forge-cd-services/idpbuilder/api/v1alpha1"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/k8s"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed resources/argo/*
var installFS embed.FS

const (
	argoApplicationControllerName string = "argocd-application-controller"
	argoServerName                string = "argocd-server"
	argoRepoServerName            string = "argocd-repo-server"
)

func GetRawInstallResources() ([][]byte, error) {
	return util.ConvertFSToBytes(installFS, "resources/argo")
}

func GetK8sInstallResources(scheme *runtime.Scheme) ([]client.Object, error) {
	rawResources, err := GetRawInstallResources()
	if err != nil {
		return nil, err
	}

	return k8s.ConvertRawResourcesToObjects(scheme, rawResources)
}

func newArgoNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}
}

func (r *LocalbuildReconciler) ReconcileArgo(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if !resource.Spec.PackageConfigs.Argo.Enabled {
		log.Info("Argo installation disabled, skipping")
		return ctrl.Result{}, nil
	}

	// Install Argo
	argonsClient := client.NewNamespacedClient(r.Client, "argocd")
	installObjs, err := GetK8sInstallResources(r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Ensure namespace exists
	argocdNs := newArgoNamespace()
	if err = r.Client.Get(ctx, types.NamespacedName{Name: "argocd"}, argocdNs); err != nil {
		// We got an error so try creating the NS
		if err = r.Client.Create(ctx, argocdNs); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("Installing argo resources")
	for _, obj := range installObjs {
		// Find the objects we need to track and own for status updates
		if obj.GetObjectKind().GroupVersionKind().Kind == "StatefulSet" && obj.GetName() == argoApplicationControllerName {
			if err = controllerutil.SetControllerReference(resource, obj, r.Scheme); err != nil {
				log.Error(err, "Setting controller reference for Argo application controller")
				return ctrl.Result{}, err
			}
		} else if obj.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
			switch obj.GetName() {
			case argoServerName:
				fallthrough
			case argoRepoServerName:
				gotObj := appsv1.Deployment{}
				if err := r.Client.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, &gotObj); err != nil {
					if err = controllerutil.SetControllerReference(resource, obj, r.Scheme); err != nil {
						log.Error(err, "Setting controller reference for Argo deployment", "deployment", obj)
						return ctrl.Result{}, err
					}
				}
			}
		}

		// Create object
		if err = k8s.EnsureObject(ctx, argonsClient, obj, "argocd"); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Set Argo available.
	// TODO(greghaynes) This should actually wait for status of some resources
	resource.Status.ArgoAvailable = true

	return ctrl.Result{}, nil
}
