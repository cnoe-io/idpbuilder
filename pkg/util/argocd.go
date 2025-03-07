package util

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/util/idp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ArgocdInitialAdminSecretName = "argocd-initial-admin-secret"
	ArgocdAdminName              = "admin"
	ArgocdNamespace              = "argocd"
	ArgocdURLTempl               = "%s://%s%s:%s%s"
)

func ArgocdBaseUrl(ctx context.Context) (string, error) {
	idpConfig, err := idp.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("error fetching idp config: %s", err)
	}
	if idpConfig.UsePathRouting {
		return fmt.Sprintf(ArgocdURLTempl, idpConfig.Protocol, "", idpConfig.Host, idpConfig.Port, "/argocd"), nil
	}
	return fmt.Sprintf(ArgocdURLTempl, idpConfig.Protocol, "argocd.", idpConfig.Host, idpConfig.Port, ""), nil
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
