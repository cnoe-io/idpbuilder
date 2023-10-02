package gitserver

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	gitServerResourceName            string = "gitserver"
	gitServerDeploymentContainerName string = "httpd"
  gitServerIngressHostnameBase     string = ".idpbuilder.cnoe.io.local"
	repoUrlFmt                       string = "http://%s.%s.svc/idpbuilder-resources.git"
)

func getRepoUrl(resource *v1alpha1.GitServer) string {
	return fmt.Sprintf(repoUrlFmt, managedResourceName(resource), resource.Namespace)
}

var gitServerLabelKey string = fmt.Sprintf("%s-gitserver", globals.ProjectName)

func ingressHostname(resource *v1alpha1.GitServer) string {
	return fmt.Sprintf("%s%s", resource.Name, gitServerIngressHostnameBase)
}

func managedResourceName(resource *v1alpha1.GitServer) string {
	return fmt.Sprintf("%s-%s", gitServerResourceName, resource.Name)
}

type subReconciler func(ctx context.Context, req ctrl.Request, resource *v1alpha1.GitServer) (ctrl.Result, error)

type GitServerReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Content fs.FS
}

func GetGitServerLabels(resource *v1alpha1.GitServer) map[string]string {
	return map[string]string{
		"app": fmt.Sprintf("%s-%s", globals.ProjectName, resource.Name),
	}
}

func SetGitDeploymentPodTemplateSpec(resource *v1alpha1.GitServer, target *appsv1.Deployment) {
	target.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: GetGitServerLabels(resource),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  gitServerDeploymentContainerName,
				Image: GetImageUrl(resource),
				Ports: []v1.ContainerPort{{
					Name:          "http",
					ContainerPort: 80,
				}},
				ReadinessProbe: &v1.Probe{
					ProbeHandler: v1.ProbeHandler{
						HTTPGet: &v1.HTTPGetAction{
							Path: "idpbuilder-resources.git/HEAD",
							Port: intstr.FromInt(80),
						},
					},
				},
			}},
		},
	}
}

func SetIngressSpec(resource *v1alpha1.GitServer, ingress *networkingv1.Ingress) {
	ingressName := "nginx"
	pathType := networkingv1.PathTypePrefix

	ingress.Spec = networkingv1.IngressSpec{
		IngressClassName: &ingressName,
		Rules: []networkingv1.IngressRule{{
			Host: ingressHostname(resource),
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{{
						Path:     "/",
						PathType: &pathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: managedResourceName(resource),
								Port: networkingv1.ServiceBackendPort{
									Name: "http",
								},
							},
						},
					}},
				},
			},
		}},
	}
}

func SetServiceSpec(resource *v1alpha1.GitServer, service *v1.Service) {
	service.Spec = v1.ServiceSpec{
		Selector: GetGitServerLabels(resource),
		Ports: []v1.ServicePort{{
			Name:       "http",
			Protocol:   v1.ProtocolTCP,
			Port:       80,
			TargetPort: intstr.FromString("http"),
		}},
	}
}

func (g *GitServerReconciler) reconcileDeployment(ctx context.Context, req ctrl.Request, resource *v1alpha1.GitServer) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      managedResourceName(resource),
			Namespace: resource.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, g.Client, deployment, func() error {
		if err := controllerutil.SetControllerReference(resource, deployment, g.Scheme); err != nil {
			log.Error(err, "Setting controller ref on git server deployment resource")
			return err
		}

		// Deployment selector is immutable so we set this value only if
		// a new object is going to be created
		if deployment.ObjectMeta.CreationTimestamp.IsZero() {
			deployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: GetGitServerLabels(resource),
			}
		}

		SetGitDeploymentPodTemplateSpec(resource, deployment)
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	resource.Status.DeploymentAvailable = deployment.Status.AvailableReplicas >= 1

	if !resource.Status.DeploymentAvailable {
		log.Info("Waiting for deployment to become available...")
		return ctrl.Result{
			RequeueAfter: time.Second * 10,
		}, nil
	}

	return ctrl.Result{}, err
}

func (g *GitServerReconciler) reconcileService(ctx context.Context, req ctrl.Request, resource *v1alpha1.GitServer) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      managedResourceName(resource),
			Namespace: resource.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, g.Client, service, func() error {
		if err := controllerutil.SetControllerReference(resource, service, g.Scheme); err != nil {
			log.Error(err, "Setting controller ref on git server service resource")
			return err
		}

		SetServiceSpec(resource, service)
		return nil
	})
	if err != nil {
		log.Error(err, "Create or update gitserver service", "resource", service)
	}
	return ctrl.Result{}, err
}

func (g *GitServerReconciler) reconcileIngress(ctx context.Context, req ctrl.Request, resource *v1alpha1.GitServer) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      managedResourceName(resource),
			Namespace: resource.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, g.Client, ingress, func() error {
		if err := controllerutil.SetControllerReference(resource, ingress, g.Scheme); err != nil {
			log.Error(err, "Setting controller ref on git server ingress resource")
			return err
		}

		SetIngressSpec(resource, ingress)
		return nil
	})
	if err != nil {
		log.Error(err, "Create or update gitserver service", "resource", ingress)
	}
	return ctrl.Result{}, err
}

func (g *GitServerReconciler) ReconcileGitServer(ctx context.Context, req ctrl.Request, resource *v1alpha1.GitServer) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Validate source
	if err := g.ValidateSource(resource); err != nil {
		log.Error(err, "Invalid image source")
		return ctrl.Result{}, err
	}

	subReconcilers := []subReconciler{
		g.reconcileGitServerImage,
		g.reconcileDeployment,
		g.reconcileService,
		g.reconcileIngress,
	}

	for _, sub := range subReconcilers {
		result, err := sub(ctx, req, resource)
		if err != nil || result.Requeue || result.RequeueAfter != 0 {
			return result, err
		}
	}

	return ctrl.Result{}, nil
}

// Responsible to updating ObservedGeneration in status
func (g *GitServerReconciler) postProcessReconcile(ctx context.Context, req ctrl.Request, resource *v1alpha1.GitServer) {
	log := log.FromContext(ctx)

	resource.Status.ObservedGeneration = resource.GetGeneration()
	if err := g.Status().Update(ctx, resource); err != nil {
		log.Error(err, "Failed to update resource status after reconcile")
	}
}

func (g *GitServerReconciler) ValidateSource(resource *v1alpha1.GitServer) error {
	if resource.Spec.Source.Embedded && resource.Spec.Source.Image != "" {
		return fmt.Errorf("cannot specify image with embedded set to true")
	}
	if !resource.Spec.Source.Embedded && resource.Spec.Source.Image == "" {
		return fmt.Errorf("must specify embedded or image source")
	}
	return nil
}

func (g *GitServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var gitServer v1alpha1.GitServer
	if err := g.Get(ctx, req.NamespacedName, &gitServer); err != nil {
		log.Error(err, "unable to fetch Resource")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Make sure we post process
	defer g.postProcessReconcile(ctx, req, &gitServer)

	return g.ReconcileGitServer(ctx, req, &gitServer)
}

func (g *GitServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GitServer{}).
		Owns(&appsv1.Deployment{}).
		Complete(g)
}
