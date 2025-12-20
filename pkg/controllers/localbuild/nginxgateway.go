package localbuild

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	nginxGatewayName = "nginx-gateway"
)

// ReconcileNginxGateway creates or updates the NginxGateway CR for the localbuild
func (r *LocalbuildReconciler) ReconcileNginxGateway(ctx context.Context, req ctrl.Request, resource *v1alpha1.Localbuild) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Reconciling NginxGateway for localbuild")

	// Create NginxGateway CR
	nginxGateway := &v1alpha2.NginxGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nginxGatewayName,
			Namespace: globals.GetProjectNamespace(resource.Name),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, nginxGateway, func() error {
		// Set controller reference
		if err := controllerutil.SetControllerReference(resource, nginxGateway, r.Scheme); err != nil {
			return fmt.Errorf("setting controller reference: %w", err)
		}

		// Set spec
		nginxGateway.Spec = v1alpha2.NginxGatewaySpec{
			Namespace: "ingress-nginx",
			Version:   "1.13.0",
			IngressClass: v1alpha2.NginxIngressClass{
				Name:      "nginx",
				IsDefault: true,
			},
		}

		return nil
	})

	if err != nil {
		logger.Error(err, "Failed to create or update NginxGateway")
		return ctrl.Result{}, err
	}

	logger.V(1).Info("NginxGateway reconciled successfully", "name", nginxGatewayName)
	return ctrl.Result{}, nil
}

// ensureNginxGatewayExists checks if NginxGateway exists and creates it if not
func (r *LocalbuildReconciler) ensureNginxGatewayExists(ctx context.Context, resource *v1alpha1.Localbuild) error {
	logger := log.FromContext(ctx)

	nginxGateway := &v1alpha2.NginxGateway{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      nginxGatewayName,
		Namespace: globals.GetProjectNamespace(resource.Name),
	}, nginxGateway)

	if err != nil && client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("checking NginxGateway existence: %w", err)
	}

	// If not found, create it
	if err != nil {
		logger.V(1).Info("NginxGateway not found, creating...")
		_, err := r.ReconcileNginxGateway(ctx, ctrl.Request{}, resource)
		return err
	}

	logger.V(1).Info("NginxGateway already exists")
	return nil
}
