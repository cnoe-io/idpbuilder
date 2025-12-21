package gitopsprovider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/localbuild"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ArgoCDProviderReconciler reconciles an ArgoCDProvider object
type ArgoCDProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch

func (r *ArgoCDProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ArgoCDProvider", "name", req.Name, "namespace", req.Namespace)

	// Fetch the ArgoCDProvider instance
	argocdProvider := &v1alpha2.ArgoCDProvider{}
	if err := r.Get(ctx, req.NamespacedName, argocdProvider); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ArgoCDProvider resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ArgoCDProvider")
		return ctrl.Result{}, err
	}

	// Set initial status if not set
	if argocdProvider.Status.Phase == "" {
		argocdProvider.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, argocdProvider); err != nil {
			logger.Error(err, "Failed to update ArgoCDProvider status")
			return ctrl.Result{}, err
		}
	}

	// Install ArgoCD
	if err := r.installArgoCD(ctx, argocdProvider); err != nil {
		logger.Error(err, "Failed to install ArgoCD")
		r.setCondition(argocdProvider, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "InstallationFailed",
			Message: fmt.Sprintf("Failed to install ArgoCD: %v", err),
		})
		argocdProvider.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, argocdProvider); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after installation failure")
		}
		return ctrl.Result{}, err
	}

	// Create admin credentials if needed
	if err := r.ensureAdminCredentials(ctx, argocdProvider); err != nil {
		logger.Error(err, "Failed to ensure admin credentials")
		return ctrl.Result{}, err
	}

	// Update status with duck-typed fields
	if err := r.updateStatus(ctx, argocdProvider); err != nil {
		logger.Error(err, "Failed to update ArgoCDProvider status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled ArgoCDProvider")
	return ctrl.Result{}, nil
}

func (r *ArgoCDProviderReconciler) installArgoCD(ctx context.Context, argocdProvider *v1alpha2.ArgoCDProvider) error {
	logger := log.FromContext(ctx)
	logger.Info("Installing ArgoCD", "namespace", argocdProvider.Spec.Namespace)

	// Ensure namespace exists
	if err := k8s.EnsureNamespace(ctx, r.Client, argocdProvider.Spec.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Load and apply embedded manifests from localbuild resources
	// Reuse the same ArgoCD installation manifests from localbuild package
	installObjs, err := k8s.BuildCustomizedObjects("", "resources/argo", localbuild.GetArgoFS(), r.Scheme, nil)
	if err != nil {
		return fmt.Errorf("failed to build argocd manifests: %w", err)
	}

	nsClient := client.NewNamespacedClient(r.Client, argocdProvider.Spec.Namespace)
	for _, obj := range installObjs {
		if err := k8s.EnsureObject(ctx, nsClient, obj, argocdProvider.Spec.Namespace); err != nil {
			return fmt.Errorf("failed to create argocd resource %s: %w", obj.GetName(), err)
		}
	}

	logger.Info("ArgoCD resources created successfully")
	return nil
}

func (r *ArgoCDProviderReconciler) ensureAdminCredentials(ctx context.Context, argocdProvider *v1alpha2.ArgoCDProvider) error {
	logger := log.FromContext(ctx)

	// Check if auto-generate is enabled
	if !argocdProvider.Spec.AdminCredentials.AutoGenerate {
		// User should provide credentials via secretRef
		if argocdProvider.Spec.AdminCredentials.SecretRef == nil {
			return fmt.Errorf("admin credentials not configured: autoGenerate is false and secretRef is nil")
		}
		return nil
	}

	// Auto-generate credentials
	secretName := "argocd-admin-secret"
	if argocdProvider.Spec.AdminCredentials.SecretRef != nil {
		secretName = argocdProvider.Spec.AdminCredentials.SecretRef.Name
	}

	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: argocdProvider.Spec.Namespace,
	}, secret)

	if err == nil {
		// Secret already exists
		logger.Info("Admin credentials secret already exists", "secret", secretName)
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to check for admin secret: %w", err)
	}

	// Generate a random password
	password, err := generateRandomPassword(16)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	// Create the secret
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: argocdProvider.Spec.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"password": password,
			"username": "admin",
		},
	}

	if err := r.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create admin secret: %w", err)
	}

	logger.Info("Created admin credentials secret", "secret", secretName)
	return nil
}

func (r *ArgoCDProviderReconciler) updateStatus(ctx context.Context, argocdProvider *v1alpha2.ArgoCDProvider) error {
	logger := log.FromContext(ctx)

	// Get the ArgoCD server service to determine endpoints
	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "argocd-server",
		Namespace: argocdProvider.Spec.Namespace,
	}, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ArgoCD server service not found yet, will retry")
			return nil
		}
		return fmt.Errorf("failed to get argocd server service: %w", err)
	}

	// Update duck-typed status fields
	argocdProvider.Status.Installed = true
	argocdProvider.Status.Phase = "Ready"
	argocdProvider.Status.Version = argocdProvider.Spec.Version

	// Set internal endpoint
	argocdProvider.Status.InternalEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local",
		svc.Name, svc.Namespace)

	// Try to get external endpoint from ingress or load balancer
	// For now, we'll set a placeholder that can be updated by the platform controller
	// based on the gateway configuration
	argocdProvider.Status.Endpoint = fmt.Sprintf("https://argocd.%s", globals.DefaultHostName)

	// Set credentials secret reference
	secretName := "argocd-admin-secret"
	if argocdProvider.Spec.AdminCredentials.SecretRef != nil {
		secretName = argocdProvider.Spec.AdminCredentials.SecretRef.Name
	}
	argocdProvider.Status.CredentialsSecretRef = &v1alpha2.SecretReference{
		Name:      secretName,
		Namespace: argocdProvider.Spec.Namespace,
		Key:       "password",
	}

	// Set Ready condition
	r.setCondition(argocdProvider, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "ArgoCDInstalled",
		Message: "ArgoCD is installed and ready",
	})

	if err := r.Status().Update(ctx, argocdProvider); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	logger.Info("Updated ArgoCDProvider status", "phase", argocdProvider.Status.Phase)
	return nil
}

func (r *ArgoCDProviderReconciler) setCondition(argocdProvider *v1alpha2.ArgoCDProvider, condition metav1.Condition) {
	condition.LastTransitionTime = metav1.Now()
	meta.SetStatusCondition(&argocdProvider.Status.Conditions, condition)
}

// generateRandomPassword generates a random password of the specified length
func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArgoCDProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.ArgoCDProvider{}).
		Complete(r)
}
