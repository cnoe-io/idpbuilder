package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/util/provider"
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

const (
	platformFinalizer  = "platform.idpbuilder.cnoe.io/finalizer"
	defaultRequeueTime = time.Second * 30
)

// PlatformReconciler reconciles a Platform object
type PlatformReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms/finalizers,verbs=update
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders,verbs=get;list;watch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Reconciling Platform", "resource", req.NamespacedName)

	// Fetch the Platform instance
	platform := &v1alpha2.Platform{}
	if err := r.Get(ctx, req.NamespacedName, platform); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Platform resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Platform")
		return ctrl.Result{}, err
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(platform, platformFinalizer) {
		controllerutil.AddFinalizer(platform, platformFinalizer)
		if err := r.Update(ctx, platform); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !platform.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, platform)
	}

	// Update phase to Pending if not set
	if platform.Status.Phase == "" {
		platform.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, platform); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Aggregate provider statuses
	allReady := true

	// Aggregate Git Providers
	gitProviderStatuses, gitReady, err := r.aggregateGitProviders(ctx, platform)
	if err != nil {
		logger.Error(err, "Failed to aggregate git providers")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}
	platform.Status.Providers.GitProviders = gitProviderStatuses
	if !gitReady {
		allReady = false
	}

	// Aggregate Gateway Providers
	gatewayStatuses, gatewayReady, err := r.aggregateGateways(ctx, platform)
	if err != nil {
		logger.Error(err, "Failed to aggregate gateways")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}
	platform.Status.Providers.Gateways = gatewayStatuses
	if !gatewayReady {
		allReady = false
	}

	// Update observed generation
	platform.Status.ObservedGeneration = platform.Generation

	// Set condition and phase based on provider readiness
	if allReady && len(platform.Spec.Components.GitProviders) > 0 {
		platform.Status.Phase = "Ready"
		meta.SetStatusCondition(&platform.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionTrue,
			Reason:  "AllComponentsReady",
			Message: "All platform components are operational",
		})
	} else {
		platform.Status.Phase = "Initializing"
		meta.SetStatusCondition(&platform.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "ComponentsNotReady",
			Message: "Waiting for platform components to be ready",
		})
	}

	if err := r.Status().Update(ctx, platform); err != nil {
		return ctrl.Result{}, err
	}

	if !allReady {
		logger.Info("Platform not fully ready, requeuing")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	logger.Info("Platform reconciliation complete", "phase", platform.Status.Phase)
	return ctrl.Result{}, nil
}

// aggregateGitProviders aggregates status from all Git providers
func (r *PlatformReconciler) aggregateGitProviders(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
	logger := log.FromContext(ctx)
	summaries := []v1alpha2.ProviderStatusSummary{}
	allReady := true

	for _, gitProviderRef := range platform.Spec.Components.GitProviders {
		// Fetch provider using unstructured client to support duck-typing
		gvk := schema.GroupVersionKind{
			Group:   "idpbuilder.cnoe.io",
			Version: "v1alpha2",
			Kind:    gitProviderRef.Kind,
		}

		providerObj := &unstructured.Unstructured{}
		providerObj.SetGroupVersionKind(gvk)

		err := r.Get(ctx, types.NamespacedName{
			Name:      gitProviderRef.Name,
			Namespace: gitProviderRef.Namespace,
		}, providerObj)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Git provider not found", "name", gitProviderRef.Name, "kind", gitProviderRef.Kind)
				summaries = append(summaries, v1alpha2.ProviderStatusSummary{
					Name:  gitProviderRef.Name,
					Kind:  gitProviderRef.Kind,
					Ready: false,
				})
				allReady = false
				continue
			}
			return nil, false, fmt.Errorf("getting git provider %s: %w", gitProviderRef.Name, err)
		}

		// Extract status using duck-typing
		ready, err := provider.IsGitProviderReady(providerObj)
		if err != nil {
			logger.Error(err, "Failed to check git provider readiness", "name", gitProviderRef.Name)
			ready = false
		}

		summaries = append(summaries, v1alpha2.ProviderStatusSummary{
			Name:  gitProviderRef.Name,
			Kind:  gitProviderRef.Kind,
			Ready: ready,
		})

		if !ready {
			allReady = false
		}
	}

	return summaries, allReady, nil
}

// aggregateGateways aggregates status from all Gateway providers
func (r *PlatformReconciler) aggregateGateways(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
	logger := log.FromContext(ctx)
	summaries := []v1alpha2.ProviderStatusSummary{}
	allReady := true

	for _, gatewayRef := range platform.Spec.Components.Gateways {
		// Fetch provider using unstructured client to support duck-typing
		gvk := schema.GroupVersionKind{
			Group:   "idpbuilder.cnoe.io",
			Version: "v1alpha2",
			Kind:    gatewayRef.Kind,
		}

		providerObj := &unstructured.Unstructured{}
		providerObj.SetGroupVersionKind(gvk)

		err := r.Get(ctx, types.NamespacedName{
			Name:      gatewayRef.Name,
			Namespace: gatewayRef.Namespace,
		}, providerObj)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Gateway provider not found", "name", gatewayRef.Name, "kind", gatewayRef.Kind)
				summaries = append(summaries, v1alpha2.ProviderStatusSummary{
					Name:  gatewayRef.Name,
					Kind:  gatewayRef.Kind,
					Ready: false,
				})
				allReady = false
				continue
			}
			return nil, false, fmt.Errorf("getting gateway provider %s: %w", gatewayRef.Name, err)
		}

		// Extract status using duck-typing
		ready, err := provider.IsGatewayProviderReady(providerObj)
		if err != nil {
			logger.Error(err, "Failed to check gateway provider readiness", "name", gatewayRef.Name)
			ready = false
		}

		summaries = append(summaries, v1alpha2.ProviderStatusSummary{
			Name:  gatewayRef.Name,
			Kind:  gatewayRef.Kind,
			Ready: ready,
		})

		if !ready {
			allReady = false
		}
	}

	return summaries, allReady, nil
}

// handleDeletion handles the deletion of Platform
func (r *PlatformReconciler) handleDeletion(ctx context.Context, platform *v1alpha2.Platform) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(platform, platformFinalizer) {
		// Perform cleanup if needed
		logger.Info("Cleaning up Platform resources")

		// Remove finalizer
		controllerutil.RemoveFinalizer(platform, platformFinalizer)
		if err := r.Update(ctx, platform); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.Platform{}).
		Complete(r)
}
