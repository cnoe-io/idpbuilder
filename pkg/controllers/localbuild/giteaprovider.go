package localbuild

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	giteaProviderName = "gitea-provider"
)

// ReconcileGiteaProvider creates or updates the GiteaProvider CR for the localbuild
func (r *LocalbuildReconciler) ReconcileGiteaProvider(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Reconciling GiteaProvider for localbuild")

	// Create GiteaProvider CR
	giteaProvider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      giteaProviderName,
			Namespace: globals.GetProjectNamespace(resource.Name),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, giteaProvider, func() error {
		// Set controller reference
		if err := controllerutil.SetControllerReference(resource, giteaProvider, r.Scheme); err != nil {
			return fmt.Errorf("setting controller reference: %w", err)
		}

		// Set spec with build customization values
		giteaProvider.Spec = v1alpha2.GiteaProviderSpec{
			Namespace:      util.GiteaNamespace,
			Version:        "1.24.3",
			Protocol:       resource.Spec.BuildCustomization.Protocol,
			Host:           resource.Spec.BuildCustomization.Host,
			Port:           resource.Spec.BuildCustomization.Port,
			UsePathRouting: resource.Spec.BuildCustomization.UsePathRouting,
			AdminUser: v1alpha2.GiteaAdminUser{
				Username:     v1alpha1.GiteaAdminUserName,
				Email:        "admin@cnoe.localtest.me",
				AutoGenerate: true,
			},
		}

		return nil
	})

	if err != nil {
		logger.Error(err, "Failed to create or update GiteaProvider")
		return ctrl.Result{}, err
	}

	logger.V(1).Info("GiteaProvider reconciled successfully", "name", giteaProviderName)
	return ctrl.Result{}, nil
}

// ensureGiteaProviderExists checks if GiteaProvider exists and creates it if not
func (r *LocalbuildReconciler) ensureGiteaProviderExists(ctx context.Context, resource *v1alpha1.Localbuild) error {
	logger := log.FromContext(ctx)

	giteaProvider := &v1alpha2.GiteaProvider{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      giteaProviderName,
		Namespace: globals.GetProjectNamespace(resource.Name),
	}, giteaProvider)

	if err != nil && client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("checking GiteaProvider existence: %w", err)
	}

	// If not found, create it
	if err != nil {
		logger.V(1).Info("GiteaProvider not found, creating...")
		_, err := r.ReconcileGiteaProvider(ctx, ctrl.Request{}, resource)
		return err
	}

	logger.V(1).Info("GiteaProvider already exists")
	return nil
}
