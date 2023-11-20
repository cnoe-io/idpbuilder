package localbuild

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var timeout = time.After(5 * time.Minute)

type EmbeddedInstallation struct {
	name         string
	resourcePath string
	namespace    string

	// skips waiting on expected resources to become ready
	skipReadinessCheck bool

	// name and gvk pair for resources that need to be monitored
	monitoredResources map[string]schema.GroupVersionKind

	resourceFS embed.FS
}

func (e *EmbeddedInstallation) rawInstallResources() ([][]byte, error) {
	return util.ConvertFSToBytes(e.resourceFS, e.resourcePath)
}

func (e *EmbeddedInstallation) installResources(scheme *runtime.Scheme) ([]client.Object, error) {
	rawResources, err := e.rawInstallResources()
	if err != nil {
		return nil, err
	}

	return k8s.ConvertRawResourcesToObjects(scheme, rawResources)
}

func (e *EmbeddedInstallation) newNamespace(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func (e *EmbeddedInstallation) Install(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild, cli client.Client, sc *runtime.Scheme) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	nsClient := client.NewNamespacedClient(cli, e.namespace)
	installObjs, err := e.installResources(sc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Ensure namespace exists
	newNS := e.newNamespace(e.namespace)
	if err = cli.Get(ctx, types.NamespacedName{Name: e.namespace}, newNS); err != nil {
		// We got an error so try creating the NS
		if err = cli.Create(ctx, newNS); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info(fmt.Sprintf("Installing/Reconciling %s resources", e.name))
	for _, obj := range installObjs {
		if gvk, ok := e.monitoredResources[obj.GetName()]; ok {
			if obj.GetObjectKind().GroupVersionKind() == gvk {
				sch := runtime.NewScheme()
				_ = appsv1.AddToScheme(sch)
				if gvkObj, err := sch.New(gvk); err == nil {
					if gotObj, ok := gvkObj.(client.Object); ok {
						if err := cli.Get(ctx, types.NamespacedName{Namespace: e.namespace, Name: obj.GetName()}, gotObj); err != nil {
							if err = controllerutil.SetControllerReference(resource, obj, sc); err != nil {
								log.Error(err, "Setting controller reference for deployment", obj.GetName(), obj)
								return ctrl.Result{}, err
							}
						}
					}
				}
			}
		}

		// Create object
		if err = k8s.EnsureObject(ctx, nsClient, obj, e.namespace); err != nil {
			return ctrl.Result{}, err
		}
	}

	// return early if readiness check is disabled
	if e.skipReadinessCheck {
		return ctrl.Result{}, nil
	}

	// wait for expected resources to become available
	errCh := make(chan error)
	var wg sync.WaitGroup

	for _, obj := range installObjs {
		if gvk, ok := e.monitoredResources[obj.GetName()]; ok {
			if obj.GetObjectKind().GroupVersionKind() != gvk {
				continue
			}

			wg.Add(1)
			go func(obj client.Object, gvk schema.GroupVersionKind) {
				defer wg.Done()

				sch := runtime.NewScheme()
				_ = appsv1.AddToScheme(sch)
				gvkObj, err := sch.New(gvk)
				if err != nil {
					errCh <- err
					return
				}

				for {
					if gotObj, ok := gvkObj.(client.Object); ok {
						if err := cli.Get(ctx, types.NamespacedName{Namespace: e.namespace, Name: obj.GetName()}, gotObj); err != nil {
							errCh <- err
							return
						}

						switch t := gotObj.(type) {
						case *appsv1.Deployment:
							if t.Status.AvailableReplicas >= 1 {
								log.Info(t.GetName(), "deployment", t.Status.AvailableReplicas)
								return
							}
						case *appsv1.StatefulSet:
							if t.Status.AvailableReplicas >= 1 {
								log.Info(t.GetName(), "statefulset", t.Status.AvailableReplicas)
								return
							}
						}
					}

					log.Info(fmt.Sprintf("Waiting for %s %s to become ready", gvk.Kind, obj.GetName()))
					time.Sleep(30 * time.Second)
				}
			}(obj, gvk)
		}
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	select {
	case <-timeout:
		err := errors.New("Timeout")
		log.Error(err, fmt.Sprintf("Didn't reconcile %s on time", e.name))
		return ctrl.Result{}, err
	case err, errOccurred := <-errCh:
		if !errOccurred {
			log.Info(fmt.Sprintf("%s is ready!", e.name))
		} else {
			log.Error(err, fmt.Sprintf("failed to reconcile the %s resources", e.name))
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
