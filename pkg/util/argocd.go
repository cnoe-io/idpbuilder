package util

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
)

const (
	ArgocdInitialAdminSecretName = "argocd-initial-admin-secret"
	ArgocdAdminName              = "admin"
	ArgocdNamespace              = "argocd"
	ArgocdIngressURL             = "%s://argocd.cnoe.localtest.me:%s"
)

func ArgocdBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	return fmt.Sprintf(ArgocdIngressURL, config.Protocol, config.Port)
}
