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

	// Reconcile Git providers
	gitProvidersSummary, allGitProvidersReady, err := r.reconcileGitProviders(ctx, platform)
	if err != nil {
		logger.Error(err, "Failed to reconcile git providers")
		meta.SetStatusCondition(&platform.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "GitProvidersFailed",
			Message: err.Error(),
		})
		platform.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, platform); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	// Update Platform status with aggregated provider status
	platform.Status.Providers.GitProviders = gitProvidersSummary
	platform.Status.ObservedGeneration = platform.Generation

	// Determine overall platform readiness
	allReady := allGitProvidersReady

	if !allReady {
		logger.V(1).Info("Not all providers are ready, requeuing")
		meta.SetStatusCondition(&platform.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "ProvidersNotReady",
			Message: "One or more providers are not ready",
		})
		platform.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, platform); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	// All providers are ready
	meta.SetStatusCondition(&platform.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "AllProvidersReady",
		Message: "All providers are ready",
	})
	platform.Status.Phase = "Ready"

	if err := r.Status().Update(ctx, platform); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Platform reconciliation complete", "phase", platform.Status.Phase)
	return ctrl.Result{}, nil
}

// reconcileGitProviders reconciles all Git provider references and returns aggregated status
func (r *PlatformReconciler) reconcileGitProviders(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
	logger := log.FromContext(ctx)
	var summary []v1alpha2.ProviderStatusSummary
	allReady := true

	for _, providerRef := range platform.Spec.Components.GitProviders {
		logger.V(1).Info("Processing git provider", "name", providerRef.Name, "kind", providerRef.Kind)

		// Fetch the provider using unstructured client for duck-typing
		gvk := schema.GroupVersionKind{
			Group:   v1alpha2.GroupVersion.Group,
			Version: v1alpha2.GroupVersion.Version,
			Kind:    providerRef.Kind,
		}

		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)

		err := r.Get(ctx, types.NamespacedName{
			Name:      providerRef.Name,
			Namespace: providerRef.Namespace,
		}, obj)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Provider not found", "name", providerRef.Name, "kind", providerRef.Kind)
				summary = append(summary, v1alpha2.ProviderStatusSummary{
					Name:  providerRef.Name,
					Kind:  providerRef.Kind,
					Ready: false,
				})
				allReady = false
				continue
			}
			return nil, false, fmt.Errorf("failed to get provider %s/%s: %w", providerRef.Kind, providerRef.Name, err)
		}

		// Use duck-typing to get provider status
		providerStatus, err := provider.GetGitProviderStatus(obj)
		if err != nil {
			logger.Error(err, "Failed to get provider status using duck-typing", "name", providerRef.Name)
			summary = append(summary, v1alpha2.ProviderStatusSummary{
				Name:  providerRef.Name,
				Kind:  providerRef.Kind,
				Ready: false,
			})
			allReady = false
			continue
		}

		summary = append(summary, v1alpha2.ProviderStatusSummary{
			Name:  providerRef.Name,
			Kind:  providerRef.Kind,
			Ready: providerStatus.Ready,
		})

		if !providerStatus.Ready {
			allReady = false
		}
	}

	return summary, allReady, nil
}

// handleDeletion handles cleanup when Platform is deleted
func (r *PlatformReconciler) handleDeletion(ctx context.Context, platform *v1alpha2.Platform) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling Platform deletion")

	// Remove finalizer
	controllerutil.RemoveFinalizer(platform, platformFinalizer)
	if err := r.Update(ctx, platform); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.Platform{}).
		Complete(r)
}
