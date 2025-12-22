package gitopsprovider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/localbuild"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "InstallationFailed",
			Message:            fmt.Sprintf("Failed to install ArgoCD: %v", err),
			LastTransitionTime: metav1.Now(),
		})
		argocdProvider.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, argocdProvider); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after installation failure")
		}
		return ctrl.Result{}, err
	}

	// Update status after installation to persist condition changes
	if err := r.Status().Update(ctx, argocdProvider); err != nil {
		logger.Error(err, "Failed to update status after installation")
		return ctrl.Result{}, err
	}

	// Create admin credentials if needed
	if err := r.ensureAdminCredentials(ctx, argocdProvider); err != nil {
		logger.Error(err, "Failed to ensure admin credentials")
		// Update status to persist condition changes
		if statusErr := r.Status().Update(ctx, argocdProvider); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after credentials failure")
		}
		return ctrl.Result{}, err
	}

	// Update status after credentials to persist condition changes
	if err := r.Status().Update(ctx, argocdProvider); err != nil {
		logger.Error(err, "Failed to update status after credentials")
		return ctrl.Result{}, err
	}

	// Check if ArgoCD is ready
	ready, err := r.isArgoCDReady(ctx, argocdProvider)
	if err != nil {
		logger.Error(err, "Failed to check ArgoCD readiness")
		// Update status to persist condition changes from isArgoCDReady
		if statusErr := r.Status().Update(ctx, argocdProvider); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after readiness check")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if !ready {
		logger.Info("ArgoCD not ready yet, requeuing")
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "Installing",
			Message:            "ArgoCD installation in progress",
			LastTransitionTime: metav1.Now(),
		})
		argocdProvider.Status.Phase = "Installing"
		if err := r.Status().Update(ctx, argocdProvider); err != nil {
			logger.Error(err, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
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
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "NamespaceReady",
			Status:  metav1.ConditionFalse,
			Reason:  "NamespaceCreationFailed",
			Message: fmt.Sprintf("Failed to ensure namespace: %v", err),
		})
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}
	meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
		Type:    "NamespaceReady",
		Status:  metav1.ConditionTrue,
		Reason:  "NamespaceExists",
		Message: "Namespace is ready",
	})

	// Retrieve template data for ArgoCD manifests
	templateData, err := r.getTemplateData(ctx)
	if err != nil {
		return fmt.Errorf("failed to get template data: %w", err)
	}

	// Load and apply embedded manifests from localbuild resources
	// Reuse the same ArgoCD installation manifests from localbuild package
	installObjs, err := k8s.BuildCustomizedObjects("", "resources/argo", localbuild.GetArgoFS(), r.Scheme, templateData)
	if err != nil {
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "ResourcesInstalled",
			Status:  metav1.ConditionFalse,
			Reason:  "ManifestBuildFailed",
			Message: fmt.Sprintf("Failed to build ArgoCD manifests: %v", err),
		})
		return fmt.Errorf("failed to build argocd manifests: %w", err)
	}

	nsClient := client.NewNamespacedClient(r.Client, argocdProvider.Spec.Namespace)
	for _, obj := range installObjs {
		if err := k8s.EnsureObject(ctx, nsClient, obj, argocdProvider.Spec.Namespace); err != nil {
			meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
				Type:    "ResourcesInstalled",
				Status:  metav1.ConditionFalse,
				Reason:  "ResourceCreationFailed",
				Message: fmt.Sprintf("Failed to create ArgoCD resource %s: %v", obj.GetName(), err),
			})
			return fmt.Errorf("failed to create argocd resource %s: %w", obj.GetName(), err)
		}
	}

	meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
		Type:    "ResourcesInstalled",
		Status:  metav1.ConditionTrue,
		Reason:  "ResourcesApplied",
		Message: "ArgoCD resources have been installed",
	})

	logger.Info("ArgoCD resources created successfully")
	return nil
}

// getTemplateData retrieves the BuildCustomizationSpec data needed for ArgoCD template rendering
func (r *ArgoCDProviderReconciler) getTemplateData(ctx context.Context) (v1alpha1.BuildCustomizationSpec, error) {
	logger := log.FromContext(ctx)

	// Retrieve the self-signed certificate from the Secret
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      globals.SelfSignedCertCMName,
		Namespace: corev1.NamespaceDefault,
	}, secret)

	var selfSignedCert string
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Self-signed certificate Secret not found, ArgoCD will be installed without TLS certificates")
			selfSignedCert = ""
		} else {
			return v1alpha1.BuildCustomizationSpec{}, fmt.Errorf("failed to get self-signed certificate: %w", err)
		}
	} else {
		certData, ok := secret.Data[globals.SelfSignedCertCMKeyName]
		if !ok {
			logger.Info("Certificate data not found in Secret, ArgoCD will be installed without TLS certificates")
			selfSignedCert = ""
		} else {
			selfSignedCert = string(certData)
		}
	}

	// Create template data with all required fields for ArgoCD templates
	templateData := v1alpha1.BuildCustomizationSpec{
		Protocol:       "https",
		Host:           globals.DefaultHostName,
		IngressHost:    globals.DefaultHostName,
		Port:           "8443",
		UsePathRouting: false,
		SelfSignedCert: selfSignedCert,
		StaticPassword: false,
	}

	return templateData, nil
}

func (r *ArgoCDProviderReconciler) isArgoCDReady(ctx context.Context, argocdProvider *v1alpha2.ArgoCDProvider) (bool, error) {
	logger := log.FromContext(ctx)

	// Check if the ArgoCD server deployment exists and is ready
	deploymentGVK := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(deploymentGVK)

	err := r.Get(ctx, types.NamespacedName{
		Namespace: argocdProvider.Spec.Namespace,
		Name:      "argocd-server",
	}, deployment)

	if err != nil {
		if apierrors.IsNotFound(err) {
			meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
				Type:    "DeploymentReady",
				Status:  metav1.ConditionFalse,
				Reason:  "DeploymentNotFound",
				Message: "ArgoCD server deployment not found",
			})
			return false, nil
		}
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "DeploymentReady",
			Status:  metav1.ConditionUnknown,
			Reason:  "DeploymentCheckFailed",
			Message: fmt.Sprintf("Failed to check deployment status: %v", err),
		})
		return false, err
	}

	// Check deployment status
	availableReplicas, found, err := unstructured.NestedInt64(deployment.Object, "status", "availableReplicas")
	if err != nil || !found {
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "DeploymentReady",
			Status:  metav1.ConditionFalse,
			Reason:  "NoAvailableReplicas",
			Message: "Deployment has no available replicas",
		})
		return false, nil
	}

	if availableReplicas < 1 {
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "DeploymentReady",
			Status:  metav1.ConditionFalse,
			Reason:  "NoAvailableReplicas",
			Message: fmt.Sprintf("Deployment has %d available replicas, need at least 1", availableReplicas),
		})
		return false, nil
	}

	meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
		Type:    "DeploymentReady",
		Status:  metav1.ConditionTrue,
		Reason:  "DeploymentAvailable",
		Message: fmt.Sprintf("Deployment has %d available replicas", availableReplicas),
	})

	// Check if ArgoCD API endpoint is accessible via the service
	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      "argocd-server",
		Namespace: argocdProvider.Spec.Namespace,
	}, svc)

	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ArgoCD server service not found yet")
			meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
				Type:    "APIAccessible",
				Status:  metav1.ConditionFalse,
				Reason:  "ServiceNotFound",
				Message: "ArgoCD server service not found",
			})
			return false, nil
		}
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "APIAccessible",
			Status:  metav1.ConditionUnknown,
			Reason:  "ServiceCheckFailed",
			Message: fmt.Sprintf("Failed to check service: %v", err),
		})
		return false, err
	}

	// If service exists and deployment is ready, consider API accessible
	// In a real cluster, we could make an HTTP call to verify, but for now
	// we'll rely on the deployment being ready
	meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
		Type:    "APIAccessible",
		Status:  metav1.ConditionTrue,
		Reason:  "ServiceReady",
		Message: "ArgoCD server service is ready",
	})

	return true, nil
}

func (r *ArgoCDProviderReconciler) ensureAdminCredentials(ctx context.Context, argocdProvider *v1alpha2.ArgoCDProvider) error {
	logger := log.FromContext(ctx)

	// Check if auto-generate is enabled
	if !argocdProvider.Spec.AdminCredentials.AutoGenerate {
		// User should provide credentials via secretRef
		if argocdProvider.Spec.AdminCredentials.SecretRef == nil {
			meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
				Type:    "AdminSecretReady",
				Status:  metav1.ConditionFalse,
				Reason:  "AdminSecretNotConfigured",
				Message: "Admin credentials not configured: autoGenerate is false and secretRef is nil",
			})
			return fmt.Errorf("admin credentials not configured: autoGenerate is false and secretRef is nil")
		}
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "AdminSecretReady",
			Status:  metav1.ConditionTrue,
			Reason:  "AdminSecretConfigured",
			Message: "Admin secret reference is configured",
		})
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
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "AdminSecretReady",
			Status:  metav1.ConditionTrue,
			Reason:  "AdminSecretExists",
			Message: "Admin secret is ready",
		})
		return nil
	}

	if !apierrors.IsNotFound(err) {
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "AdminSecretReady",
			Status:  metav1.ConditionFalse,
			Reason:  "AdminSecretCheckFailed",
			Message: fmt.Sprintf("Failed to check for admin secret: %v", err),
		})
		return fmt.Errorf("failed to check for admin secret: %w", err)
	}

	// Generate a random password
	password, err := generateRandomPassword(16)
	if err != nil {
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "AdminSecretReady",
			Status:  metav1.ConditionFalse,
			Reason:  "PasswordGenerationFailed",
			Message: fmt.Sprintf("Failed to generate password: %v", err),
		})
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
		meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
			Type:    "AdminSecretReady",
			Status:  metav1.ConditionFalse,
			Reason:  "AdminSecretCreationFailed",
			Message: fmt.Sprintf("Failed to create admin secret: %v", err),
		})
		return fmt.Errorf("failed to create admin secret: %w", err)
	}

	meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
		Type:    "AdminSecretReady",
		Status:  metav1.ConditionTrue,
		Reason:  "AdminSecretCreated",
		Message: "Admin secret is ready",
	})

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
	meta.SetStatusCondition(&argocdProvider.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "ArgoCDInstalled",
		Message:            "ArgoCD is installed and ready",
		LastTransitionTime: metav1.Now(),
	})

	if err := r.Status().Update(ctx, argocdProvider); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	logger.Info("Updated ArgoCDProvider status", "phase", argocdProvider.Status.Phase)
	return nil
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
