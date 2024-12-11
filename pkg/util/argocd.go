package util

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
)

const (
	ArgocdDevModePassword        = "developer"
	ArgocdInitialAdminSecretName = "argocd-initial-admin-secret"
	ArgocdNamespace              = "argocd"
	ArgocdIngressURL             = "%s://argocd.cnoe.localtest.me:%s"
)

func ArgocdBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	return fmt.Sprintf(ArgocdIngressURL, config.Protocol, config.Port)
}
