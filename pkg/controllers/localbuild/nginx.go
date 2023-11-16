package localbuild

import (
	"context"
	"embed"
	"errors"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
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

const (
	nginxNamespace  string = "ingress-nginx"
	nginxDeployment string = "ingress-nginx-controller"
)

//go:embed resources/nginx/k8s/*
var installNginxFS embed.FS
var timeout = time.After(3 * time.Minute)

func RawNginxInstallResources() ([][]byte, error) {
	return util.ConvertFSToBytes(installNginxFS, "resources/nginx/k8s")
}

func NginxInstallResources(scheme *runtime.Scheme) ([]client.Object, error) {
	rawResources, err := RawNginxInstallResources()
	if err != nil {
		return nil, err
	}

	return k8s.ConvertRawResourcesToObjects(scheme, rawResources)
}

func newNginxNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nginxNamespace,
		},
	}
}

func (r LocalbuildReconciler) ReconcileNginx(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	nginxNSClient := client.NewNamespacedClient(r.Client, nginxNamespace)
	installObjs, err := NginxInstallResources(r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Ensure namespace exists
	nginxNewNS := newNginxNamespace()
	if err = r.Client.Get(ctx, types.NamespacedName{Name: nginxNamespace}, nginxNewNS); err != nil {
		// We got an error so try creating the NS
		if err = r.Client.Create(ctx, nginxNewNS); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("Installing/Reconciling Nginx resources")
	for _, obj := range installObjs {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
			switch obj.GetName() {
			case nginxDeployment:
				gotObj := appsv1.Deployment{}
				if err := r.Client.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, &gotObj); err != nil {
					if err = controllerutil.SetControllerReference(resource, obj, r.Scheme); err != nil {
						log.Error(err, "Setting controller reference for Nginx deployment", "deployment", obj)
						return ctrl.Result{}, err
					}
				}
			}
		}

		// Create object
		if err = k8s.EnsureObject(ctx, nginxNSClient, obj, nginxNamespace); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Wait for Nginx to become available
	ready := make(chan error)
	go func([]client.Object) {
		for {
			for _, obj := range installObjs {
				if obj.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
					switch obj.GetName() {
					case nginxDeployment:
						gotObj := appsv1.Deployment{}
						if err := r.Client.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, &gotObj); err != nil {
							ready <- err
							return
						}

						if gotObj.Status.AvailableReplicas >= 1 {
							close(ready)
							return
						}
					}
				}
			}
			log.Info("Waiting for Nginx to become ready")
			time.Sleep(30 * time.Second)
		}
	}(installObjs)

	select {
	case <-timeout:
		err := errors.New("Timeout")
		log.Error(err, "Didn't reconcile Nginx on time.")
		return ctrl.Result{}, err
	case err, errOccurred := <-ready:
		if !errOccurred {
			log.Info("Nginx is ready!")
			resource.Status.NginxAvailable = true
		} else {
			log.Error(err, "failed to reconcile the Nginx resources")
			resource.Status.NginxAvailable = false
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
