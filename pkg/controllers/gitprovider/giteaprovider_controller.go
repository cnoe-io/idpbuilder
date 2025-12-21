package gitprovider

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/resources/gitea"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	corev1 "k8s.io/api/core/v1"
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
	giteaProviderFinalizer           = "giteaprovider.idpbuilder.cnoe.io/finalizer"
	defaultRequeueTime               = time.Second * 30
	giteaDeploymentName              = "my-gitea"
	nginxAdmissionWebhookServiceName = "ingress-nginx-controller-admission"
	nginxNamespace                   = "ingress-nginx"
)

// GiteaProviderReconciler reconciles a GiteaProvider object
type GiteaProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config v1alpha1.BuildCustomizationSpec
}

//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *GiteaProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Reconciling GiteaProvider", "resource", req.NamespacedName)

	// Fetch the GiteaProvider instance
	provider := &v1alpha2.GiteaProvider{}
	if err := r.Get(ctx, req.NamespacedName, provider); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("GiteaProvider resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get GiteaProvider")
		return ctrl.Result{}, err
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(provider, giteaProviderFinalizer) {
		controllerutil.AddFinalizer(provider, giteaProviderFinalizer)
		if err := r.Update(ctx, provider); err != nil {
			// Conflict errors are expected when the resource is updated by another process
			// Return the error to trigger a retry without logging it as a failure
			if errors.IsConflict(err) {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !provider.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, provider)
	}

	// Update phase to Installing if not set
	if provider.Status.Phase == "" {
		provider.Status.Phase = "Installing"
		if err := r.Status().Update(ctx, provider); err != nil {
			// Conflict errors are expected when the resource is updated by another process
			// Return the error to trigger a retry without logging it as a failure
			if errors.IsConflict(err) {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	}

	// Reconcile Gitea installation
	result, err := r.reconcileGitea(ctx, provider)
	if err != nil {
		// Set condition to False
		meta.SetStatusCondition(&provider.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "InstallationFailed",
			Message: err.Error(),
		})
		provider.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, provider); statusErr != nil {
			// Don't log conflict errors as failures - they will be retried automatically
			if !errors.IsConflict(statusErr) {
				logger.Error(statusErr, "Failed to update status")
			}
		}
		return result, err
	}

	// Check if Gitea is ready
	ready, err := r.isGiteaReady(ctx, provider)
	if err != nil {
		logger.Error(err, "Failed to check Gitea readiness")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	if !ready {
		logger.Info("Gitea not ready yet, requeuing")
		meta.SetStatusCondition(&provider.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "Installing",
			Message: "Gitea installation in progress",
		})
		provider.Status.Phase = "Installing"
		if err := r.Status().Update(ctx, provider); err != nil {
			// Conflict errors are expected when the resource is updated by another process
			// Return the error to trigger a retry without logging it as a failure
			if errors.IsConflict(err) {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	// Gitea is ready, update status
	baseUrl := util.GiteaBaseUrl(r.Config)
	// Construct internal URL for cluster-internal access
	internalUrl := fmt.Sprintf("http://my-gitea-http.%s.svc.cluster.local:3000", provider.Spec.Namespace)

	// Ensure admin secret and token
	secret, err := r.ensureAdminSecret(ctx, provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update status with duck-typed fields
	provider.Status.Endpoint = baseUrl
	provider.Status.InternalEndpoint = internalUrl
	provider.Status.CredentialsSecretRef = &v1alpha2.SecretReference{
		Name:      secret.Name,
		Namespace: secret.Namespace,
		Key:       util.GiteaAdminTokenFieldName,
	}
	provider.Status.Installed = true
	provider.Status.Version = provider.Spec.Version
	provider.Status.Phase = "Ready"
	provider.Status.AdminUser.Username = provider.Spec.AdminUser.Username
	if provider.Status.AdminUser.Username == "" {
		provider.Status.AdminUser.Username = "giteaAdmin"
	}
	provider.Status.AdminUser.SecretRef = &v1alpha2.SecretReference{
		Name:      secret.Name,
		Namespace: secret.Namespace,
		Key:       "password",
	}

	// Set Ready condition
	meta.SetStatusCondition(&provider.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "GiteaReady",
		Message: "Gitea is ready and accessible",
	})

	if err := r.Status().Update(ctx, provider); err != nil {
		// Conflict errors are expected when the resource is updated by another process
		// Return the error to trigger a retry without logging it as a failure
		if errors.IsConflict(err) {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	logger.Info("GiteaProvider reconciliation complete", "endpoint", baseUrl)
	return ctrl.Result{}, nil
}

// reconcileGitea handles the installation and configuration of Gitea
func (r *GiteaProviderReconciler) reconcileGitea(ctx context.Context, provider *v1alpha2.GiteaProvider) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Ensure namespace exists
	if err := k8s.EnsureNamespace(ctx, r.Client, provider.Spec.Namespace); err != nil {
		return ctrl.Result{}, fmt.Errorf("ensuring namespace: %w", err)
	}

	// Ensure admin secret exists BEFORE installing Gitea resources
	// The deployment references this secret in environment variables
	if err := r.ensureAdminSecretWithoutToken(ctx, provider); err != nil {
		return ctrl.Result{}, fmt.Errorf("ensuring admin secret: %w", err)
	}

	// Check if nginx admission webhook is ready before creating Ingress resources
	// This prevents race condition where Gitea tries to create Ingress before webhook is available
	ready, err := r.isNginxAdmissionWebhookReady(ctx)
	if err != nil {
		logger.V(1).Info("Failed to check nginx admission webhook readiness", "error", err)
		// Continue anyway - webhook might not be installed, or this might be a different setup
	} else if !ready {
		logger.Info("Nginx admission webhook not ready yet, requeuing to prevent race condition")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	// Install Gitea resources using embedded manifests
	if err := r.installGiteaResources(ctx, provider); err != nil {
		return ctrl.Result{}, fmt.Errorf("installing Gitea resources: %w", err)
	}

	logger.V(1).Info("Gitea resources installed", "namespace", provider.Spec.Namespace)
	return ctrl.Result{}, nil
}

// installGiteaResources installs Gitea using embedded manifests from gitea package
func (r *GiteaProviderReconciler) installGiteaResources(ctx context.Context, provider *v1alpha2.GiteaProvider) error {
	logger := log.FromContext(ctx)

	// Build configuration from provider spec with defaults
	config := r.buildConfigFromSpec(provider)
	logger.V(1).Info("Using configuration for Gitea templates", "protocol", config.Protocol, "host", config.Host, "port", config.Port, "usePathRouting", config.UsePathRouting)

	// Use the exported function from gitea package to get raw Gitea resources
	rawResources, err := gitea.RawGiteaInstallResources(config, v1alpha1.PackageCustomization{}, r.Scheme)
	if err != nil {
		return fmt.Errorf("getting Gitea manifests: %w", err)
	}

	// Convert raw bytes to objects
	installObjs, err := k8s.ConvertRawResourcesToObjects(r.Scheme, rawResources)
	if err != nil {
		return fmt.Errorf("converting YAML to objects: %w", err)
	}

	nsClient := client.NewNamespacedClient(r.Client, provider.Spec.Namespace)

	for _, obj := range installObjs {
		if err := k8s.EnsureObject(ctx, nsClient, obj, provider.Spec.Namespace); err != nil {
			return fmt.Errorf("ensuring object %s: %w", obj.GetName(), err)
		}
	}

	logger.V(1).Info("Gitea manifests applied", "count", len(installObjs))
	return nil
}

// buildConfigFromSpec creates a BuildCustomizationSpec from the GiteaProvider spec with defaults
func (r *GiteaProviderReconciler) buildConfigFromSpec(provider *v1alpha2.GiteaProvider) v1alpha1.BuildCustomizationSpec {
	config := v1alpha1.BuildCustomizationSpec{
		Protocol:       provider.Spec.Protocol,
		Host:           provider.Spec.Host,
		Port:           provider.Spec.Port,
		UsePathRouting: provider.Spec.UsePathRouting,
	}

	// Apply defaults if values are empty
	if config.Protocol == "" {
		config.Protocol = "http"
	}
	if config.Host == "" {
		config.Host = "cnoe.localtest.me"
	}
	if config.Port == "" {
		config.Port = "8080"
	}
	// Set IngressHost to Host if not explicitly set
	if config.IngressHost == "" {
		config.IngressHost = config.Host
	}

	return config
}

// isGiteaReady checks if the Gitea deployment is ready
func (r *GiteaProviderReconciler) isGiteaReady(ctx context.Context, provider *v1alpha2.GiteaProvider) (bool, error) {
	logger := log.FromContext(ctx)

	// Check if deployment exists and is ready
	deploymentGVK := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}

	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(deploymentGVK)

	err := r.Get(ctx, types.NamespacedName{
		Namespace: provider.Spec.Namespace,
		Name:      giteaDeploymentName,
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

	if availableReplicas < 1 {
		return false, nil
	}

	// Check if Gitea API endpoint is accessible
	config := r.buildConfigFromSpec(provider)
	baseUrl := util.GiteaBaseUrl(config)
	logger.V(1).Info("checking gitea api endpoint", "url", baseUrl)

	c := util.GetHttpClient()
	resp, err := c.Get(baseUrl)
	if err != nil {
		logger.V(1).Info("Gitea API not yet accessible", "error", err)
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.V(1).Info("Gitea API returned non-OK status", "statusCode", resp.StatusCode)
		return false, nil
	}

	return true, nil
}

// createAdminSecretIfNotExists creates the admin secret with username and password if it doesn't exist
func (r *GiteaProviderReconciler) createAdminSecretIfNotExists(ctx context.Context, provider *v1alpha2.GiteaProvider) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)

	secretName := util.GiteaAdminSecret
	secretNamespace := provider.Spec.Namespace

	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: secretNamespace,
		Name:      secretName,
	}, secret)

	if err != nil {
		if errors.IsNotFound(err) {
			// Create new secret with generated password
			genPassword, err := util.GeneratePassword()
			if err != nil {
				return nil, fmt.Errorf("generating password: %w", err)
			}

			username := provider.Spec.AdminUser.Username
			if username == "" {
				username = "giteaAdmin"
			}

			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: secretNamespace,
				},
				StringData: map[string]string{
					"username": username,
					"password": genPassword,
				},
			}

			if err := r.Create(ctx, secret); err != nil {
				return nil, fmt.Errorf("creating admin secret: %w", err)
			}
			logger.Info("Created Gitea admin secret", "name", secretName)
		} else {
			return nil, fmt.Errorf("getting admin secret: %w", err)
		}
	}

	return secret, nil
}

// ensureAdminSecret ensures the admin secret exists and has a token
func (r *GiteaProviderReconciler) ensureAdminSecret(ctx context.Context, provider *v1alpha2.GiteaProvider) (*corev1.Secret, error) {
	// First ensure the secret exists with username and password
	secret, err := r.createAdminSecretIfNotExists(ctx, provider)
	if err != nil {
		return nil, err
	}

	// Ensure token exists
	if _, ok := secret.Data[util.GiteaAdminTokenFieldName]; !ok {
		// Get token from Gitea API
		username := string(secret.Data["username"])
		password := string(secret.Data["password"])
		baseUrl := util.GiteaBaseUrl(r.Config)

		token, err := util.GetGiteaToken(ctx, baseUrl, username, password)
		if err != nil {
			return nil, fmt.Errorf("getting Gitea token: %w", err)
		}

		// Update secret with token using base64 encoding
		encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

		// Update secret with token
		u := &unstructured.Unstructured{}
		u.SetName(secret.Name)
		u.SetNamespace(secret.Namespace)
		u.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

		if err := unstructured.SetNestedField(u.Object, encodedToken, "data", util.GiteaAdminTokenFieldName); err != nil {
			return nil, fmt.Errorf("setting token field: %w", err)
		}

		if err := r.Patch(ctx, u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager)); err != nil {
			return nil, fmt.Errorf("patching secret with token: %w", err)
		}

		// Refetch secret
		if err := r.Get(ctx, types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}, secret); err != nil {
			return nil, err
		}
	}

	return secret, nil
}

// ensureAdminSecretWithoutToken creates the admin secret with username and password only
// This should be called BEFORE installing Gitea resources since the deployment references this secret
func (r *GiteaProviderReconciler) ensureAdminSecretWithoutToken(ctx context.Context, provider *v1alpha2.GiteaProvider) error {
	_, err := r.createAdminSecretIfNotExists(ctx, provider)
	return err
}

// handleDeletion handles cleanup when GiteaProvider is deleted
func (r *GiteaProviderReconciler) handleDeletion(ctx context.Context, provider *v1alpha2.GiteaProvider) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling GiteaProvider deletion")

	// Remove finalizer
	controllerutil.RemoveFinalizer(provider, giteaProviderFinalizer)
	if err := r.Update(ctx, provider); err != nil {
		// Conflict errors are expected when the resource is updated by another process
		// Return the error to trigger a retry without logging it as a failure
		if errors.IsConflict(err) {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// isNginxAdmissionWebhookReady checks if the nginx admission webhook service is ready
// This prevents race conditions when creating Ingress resources before the webhook is available
func (r *GiteaProviderReconciler) isNginxAdmissionWebhookReady(ctx context.Context) (bool, error) {
	logger := log.FromContext(ctx)

	// Check if the nginx admission webhook service exists
	service := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: nginxNamespace,
		Name:      nginxAdmissionWebhookServiceName,
	}, service)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("Nginx admission webhook service not found")
			return false, nil
		}
		return false, err
	}

	// Check if the service has endpoints (meaning the webhook pod is ready)
	endpoints := &corev1.Endpoints{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: nginxNamespace,
		Name:      nginxAdmissionWebhookServiceName,
	}, endpoints)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("Nginx admission webhook endpoints not found")
			return false, nil
		}
		return false, err
	}

	// Check if endpoints has at least one ready address
	if len(endpoints.Subsets) == 0 {
		logger.V(1).Info("Nginx admission webhook has no endpoint subsets")
		return false, nil
	}

	// Check each subset for ready addresses
	for _, subset := range endpoints.Subsets {
		if len(subset.Addresses) > 0 {
			logger.V(1).Info("Nginx admission webhook is ready", "readyAddresses", len(subset.Addresses))
			return true, nil
		}
	}

	logger.V(1).Info("Nginx admission webhook has no ready addresses")
	return false, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *GiteaProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.GiteaProvider{}).
		Complete(r)
}
