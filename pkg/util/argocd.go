package util

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/globals"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ArgocdInitialAdminSecretName = "argocd-initial-admin-secret"
	ArgocdAdminName              = "admin"
	ArgocdNamespace              = "argocd"
	ArgocdIngressURL             = "%s://argocd.cnoe.localtest.me:%s"
)

func ArgocdBaseUrl() string {
	return fmt.Sprintf(ArgocdIngressURL, globals.ClusterProtocol, globals.ClusterPort)
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
