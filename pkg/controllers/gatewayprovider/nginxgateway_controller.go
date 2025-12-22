package gatewayprovider

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed resources/nginx/k8s/*
var installNginxFS embed.FS

const (
	nginxGatewayFinalizer      = "nginxgateway.idpbuilder.cnoe.io/finalizer"
	defaultRequeueTime         = time.Second * 30
	nginxControllerDeployment  = "ingress-nginx-controller"
	nginxControllerServiceName = "ingress-nginx-controller"
)

// NginxGatewayReconciler reconciles a NginxGateway object
type NginxGatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config v1alpha1.BuildCustomizationSpec
}

//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *NginxGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Reconciling NginxGateway", "resource", req.NamespacedName)

	// Fetch the NginxGateway instance
	gateway := &v1alpha2.NginxGateway{}
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("NginxGateway resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get NginxGateway")
		return ctrl.Result{}, err
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(gateway, nginxGatewayFinalizer) {
		controllerutil.AddFinalizer(gateway, nginxGatewayFinalizer)
		if err := r.Update(ctx, gateway); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !gateway.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, gateway)
	}

	// Update phase to Installing if not set
	if gateway.Status.Phase == "" {
		gateway.Status.Phase = "Installing"
		if err := r.Status().Update(ctx, gateway); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile Nginx installation
	result, err := r.reconcileNginx(ctx, gateway)
	if err != nil {
		// Set condition to False
		r.setCondition(gateway, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "InstallationFailed",
			Message: err.Error(),
		})
		gateway.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, gateway); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}
		return result, err
	}

	// Check if Nginx deployment is ready
	deploymentReady, err := r.isNginxReady(ctx, gateway)
	if err != nil {
		logger.Error(err, "Failed to check Nginx deployment readiness")
		r.setCondition(gateway, metav1.Condition{
			Type:    "DeploymentReady",
			Status:  metav1.ConditionUnknown,
			Reason:  "DeploymentCheckFailed",
			Message: fmt.Sprintf("Failed to check deployment status: %v", err),
		})
	} else if !deploymentReady {
		logger.Info("Nginx deployment not ready yet, requeuing")
		r.setCondition(gateway, metav1.Condition{
			Type:    "DeploymentReady",
			Status:  metav1.ConditionFalse,
			Reason:  "DeploymentNotReady",
			Message: "Nginx controller deployment is not ready yet",
		})
		r.setCondition(gateway, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "Installing",
			Message: "Nginx installation in progress",
		})
		gateway.Status.Phase = "Installing"
		if err := r.Status().Update(ctx, gateway); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	} else {
		r.setCondition(gateway, metav1.Condition{
			Type:    "DeploymentReady",
			Status:  metav1.ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "Nginx controller deployment is ready",
		})
	}

	// Check if service is ready
	serviceReady, serviceErr := r.isServiceReady(ctx, gateway)
	if serviceErr != nil {
		logger.Error(serviceErr, "Failed to check service readiness")
		r.setCondition(gateway, metav1.Condition{
			Type:    "ServiceReady",
			Status:  metav1.ConditionUnknown,
			Reason:  "ServiceCheckFailed",
			Message: fmt.Sprintf("Failed to check service status: %v", serviceErr),
		})
	} else if !serviceReady {
		logger.Info("Service not ready yet, requeuing")
		r.setCondition(gateway, metav1.Condition{
			Type:    "ServiceReady",
			Status:  metav1.ConditionFalse,
			Reason:  "ServiceNotReady",
			Message: "Nginx controller service is not ready yet",
		})
		r.setCondition(gateway, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "Installing",
			Message: "Nginx installation in progress",
		})
		gateway.Status.Phase = "Installing"
		if err := r.Status().Update(ctx, gateway); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	} else {
		r.setCondition(gateway, metav1.Condition{
			Type:    "ServiceReady",
			Status:  metav1.ConditionTrue,
			Reason:  "ServiceReady",
			Message: "Nginx controller service is ready",
		})
	}

	// Nginx is ready, update status with duck-typed fields
	ingressClass := gateway.Spec.IngressClass.Name
	if ingressClass == "" {
		ingressClass = "nginx"
	}

	// Get load balancer endpoint
	loadBalancerEndpoint, err := r.getLoadBalancerEndpoint(ctx, gateway)
	if err != nil {
		logger.Error(err, "Failed to get load balancer endpoint")
		loadBalancerEndpoint = "" // Continue without endpoint
	}

	// Construct internal endpoint
	internalEndpoint := fmt.Sprintf("http://%s.%s.svc.cluster.local", nginxControllerServiceName, gateway.Spec.Namespace)

	// Get controller replica counts
	replicas, readyReplicas, err := r.getControllerStatus(ctx, gateway)
	if err != nil {
		logger.Error(err, "Failed to get controller status")
	}

	// Update status with duck-typed fields
	gateway.Status.IngressClassName = ingressClass
	gateway.Status.LoadBalancerEndpoint = loadBalancerEndpoint
	gateway.Status.InternalEndpoint = internalEndpoint
	gateway.Status.Installed = true
	gateway.Status.Version = gateway.Spec.Version
	gateway.Status.Phase = "Ready"
	gateway.Status.Controller.Replicas = replicas
	gateway.Status.Controller.ReadyReplicas = readyReplicas

	// Set Ready condition
	r.setCondition(gateway, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "NginxReady",
		Message: "Nginx Ingress Controller is ready and accessible",
	})

	if err := r.Status().Update(ctx, gateway); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("NginxGateway reconciliation complete", "ingressClass", ingressClass, "endpoint", loadBalancerEndpoint)
	return ctrl.Result{}, nil
}

// reconcileNginx handles the installation and configuration of Nginx
func (r *NginxGatewayReconciler) reconcileNginx(ctx context.Context, gateway *v1alpha2.NginxGateway) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Ensure namespace exists
	if err := k8s.EnsureNamespace(ctx, r.Client, gateway.Spec.Namespace); err != nil {
		r.setCondition(gateway, metav1.Condition{
			Type:    "NamespaceReady",
			Status:  metav1.ConditionFalse,
			Reason:  "NamespaceCreationFailed",
			Message: fmt.Sprintf("Failed to create namespace: %v", err),
		})
		return ctrl.Result{}, fmt.Errorf("ensuring namespace: %w", err)
	}

	// Set NamespaceReady condition
	r.setCondition(gateway, metav1.Condition{
		Type:    "NamespaceReady",
		Status:  metav1.ConditionTrue,
		Reason:  "NamespaceCreated",
		Message: fmt.Sprintf("Namespace %s is ready", gateway.Spec.Namespace),
	})

	// Install Nginx resources using embedded manifests
	if err := r.installNginxResources(ctx, gateway); err != nil {
		r.setCondition(gateway, metav1.Condition{
			Type:    "ResourcesInstalled",
			Status:  metav1.ConditionFalse,
			Reason:  "ResourceInstallationFailed",
			Message: fmt.Sprintf("Failed to install Nginx resources: %v", err),
		})
		return ctrl.Result{}, fmt.Errorf("installing Nginx resources: %w", err)
	}

	// Set ResourcesInstalled condition
	r.setCondition(gateway, metav1.Condition{
		Type:    "ResourcesInstalled",
		Status:  metav1.ConditionTrue,
		Reason:  "ResourcesInstalled",
		Message: "Nginx resources have been installed successfully",
	})

	logger.V(1).Info("Nginx resources installed", "namespace", gateway.Spec.Namespace)
	return ctrl.Result{}, nil
}

// installNginxResources installs Nginx using embedded manifests
func (r *NginxGatewayReconciler) installNginxResources(ctx context.Context, gateway *v1alpha2.NginxGateway) error {
	logger := log.FromContext(ctx)

	// Use embedded resources from this package
	rawResources, err := k8s.BuildCustomizedManifests("", "resources/nginx/k8s", installNginxFS, r.Scheme, r.Config)
	if err != nil {
		return fmt.Errorf("getting Nginx manifests: %w", err)
	}

	// Convert raw bytes to objects
	installObjs, err := k8s.ConvertRawResourcesToObjects(r.Scheme, rawResources)
	if err != nil {
		return fmt.Errorf("converting YAML to objects: %w", err)
	}

	nsClient := client.NewNamespacedClient(r.Client, gateway.Spec.Namespace)

	for _, obj := range installObjs {
		if err := k8s.EnsureObject(ctx, nsClient, obj, gateway.Spec.Namespace); err != nil {
			return fmt.Errorf("ensuring object %s: %w", obj.GetName(), err)
		}
	}

	logger.V(1).Info("Nginx manifests applied", "count", len(installObjs))
	return nil
}

// isNginxReady checks if the Nginx deployment is ready
func (r *NginxGatewayReconciler) isNginxReady(ctx context.Context, gateway *v1alpha2.NginxGateway) (bool, error) {
	// Check if deployment exists and is ready
	deploymentGVK := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(deploymentGVK)

	err := r.Get(ctx, types.NamespacedName{
		Namespace: gateway.Spec.Namespace,
		Name:      nginxControllerDeployment,
	}, deployment)

	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	// Check deployment status
	availableReplicas, found, err := unstructured.NestedInt64(deployment.Object, "status", "availableReplicas")
	if err != nil || !found {
		return false, nil
	}

	replicas, found, err := unstructured.NestedInt64(deployment.Object, "status", "replicas")
	if err != nil || !found {
		return false, nil
	}

	// Ready when all replicas are available
	return availableReplicas > 0 && availableReplicas == replicas, nil
}

// getLoadBalancerEndpoint retrieves the load balancer endpoint from the Nginx service
func (r *NginxGatewayReconciler) getLoadBalancerEndpoint(ctx context.Context, gateway *v1alpha2.NginxGateway) (string, error) {
	// For Kind clusters and local development, we typically use NodePort
	// Try to get the node IP
	if r.Config.Host != "" {
		return r.Config.Host, nil
	}

	// Try to get from service
	serviceGVK := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	}

	service := &unstructured.Unstructured{}
	service.SetGroupVersionKind(serviceGVK)

	err := r.Get(ctx, types.NamespacedName{
		Namespace: gateway.Spec.Namespace,
		Name:      nginxControllerServiceName,
	}, service)

	if err != nil {
		return "", err
	}

	// Try to get LoadBalancer IP
	ingress, found, err := unstructured.NestedSlice(service.Object, "status", "loadBalancer", "ingress")
	if err == nil && found && len(ingress) > 0 {
		if ingressMap, ok := ingress[0].(map[string]interface{}); ok {
			if ip, ok := ingressMap["ip"].(string); ok && ip != "" {
				return fmt.Sprintf("http://%s", ip), nil
			}
			if hostname, ok := ingressMap["hostname"].(string); ok && hostname != "" {
				return fmt.Sprintf("http://%s", hostname), nil
			}
		}
	}

	// For NodePort or ClusterIP services, return the cluster IP
	clusterIP, found, err := unstructured.NestedString(service.Object, "spec", "clusterIP")
	if err == nil && found && clusterIP != "" && clusterIP != "None" {
		return fmt.Sprintf("http://%s", clusterIP), nil
	}

	return "", fmt.Errorf("could not determine load balancer endpoint")
}

// getControllerStatus retrieves the controller deployment status
func (r *NginxGatewayReconciler) getControllerStatus(ctx context.Context, gateway *v1alpha2.NginxGateway) (int32, int32, error) {
	deploymentGVK := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(deploymentGVK)

	err := r.Get(ctx, types.NamespacedName{
		Namespace: gateway.Spec.Namespace,
		Name:      nginxControllerDeployment,
	}, deployment)

	if err != nil {
		return 0, 0, err
	}

	replicas, _, _ := unstructured.NestedInt64(deployment.Object, "status", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(deployment.Object, "status", "readyReplicas")

	return int32(replicas), int32(readyReplicas), nil
}

// isServiceReady checks if the Nginx service exists and is ready
func (r *NginxGatewayReconciler) isServiceReady(ctx context.Context, gateway *v1alpha2.NginxGateway) (bool, error) {
	serviceGVK := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	}

	service := &unstructured.Unstructured{}
	service.SetGroupVersionKind(serviceGVK)

	err := r.Get(ctx, types.NamespacedName{
		Namespace: gateway.Spec.Namespace,
		Name:      nginxControllerServiceName,
	}, service)

	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	// Service exists, consider it ready
	// Note: We check only for service existence rather than endpoints because:
	// 1. The deployment readiness check already validates that pods are running
	// 2. Service endpoints will be automatically populated once pods are ready
	// 3. This simplifies the check while still providing meaningful status
	return true, nil
}

// handleDeletion handles the deletion of NginxGateway
func (r *NginxGatewayReconciler) handleDeletion(ctx context.Context, gateway *v1alpha2.NginxGateway) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(gateway, nginxGatewayFinalizer) {
		// Perform cleanup if needed
		logger.Info("Cleaning up NginxGateway resources", "namespace", gateway.Spec.Namespace)

		// Remove finalizer
		controllerutil.RemoveFinalizer(gateway, nginxGatewayFinalizer)
		if err := r.Update(ctx, gateway); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// setCondition sets a condition on the NginxGateway status with the current timestamp
func (r *NginxGatewayReconciler) setCondition(gateway *v1alpha2.NginxGateway, condition metav1.Condition) {
	condition.LastTransitionTime = metav1.Now()
	meta.SetStatusCondition(&gateway.Status.Conditions, condition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.NginxGateway{}).
		Complete(r)
}
