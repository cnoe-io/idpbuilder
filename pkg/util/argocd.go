package util

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ArgocdInitialAdminSecretName = "argocd-initial-admin-secret"
	ArgocdAdminName              = "admin"
	ArgocdNamespace              = "argocd"
	ArgocdURLTempl               = "%s://%s%s:%s%s"
)

func ArgocdBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	if config.UsePathRouting {
		return fmt.Sprintf(ArgocdURLTempl, config.Protocol, "", config.Host, config.Port, "/argocd")
	}
	return fmt.Sprintf(ArgocdURLTempl, config.Protocol, "argocd.", config.Host, config.Port, "")
}

func ArgocdInitialAdminSecretObject() corev1.Secret {
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ArgocdInitialAdminSecretName,
			Namespace: ArgocdNamespace,
		},
	}
}
