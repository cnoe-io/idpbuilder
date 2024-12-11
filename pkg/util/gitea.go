package util

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// hardcoded values from what we have in the yaml installation file.
	GiteaNamespace           = "gitea"
	GiteaAdminSecret         = "gitea-credential"
	GiteaAdminTokenName      = "admin"
	GiteaAdminTokenFieldName = "token"
	// this is the URL accessible outside cluster. resolves to localhost
	GiteaIngressURL = "%s://gitea.cnoe.localtest.me:%s"
	// this is the URL accessible within cluster for ArgoCD to fetch resources.
	// resolves to cluster ip
	GiteaSvcURL = "%s://%s%s:%s%s"
)

func GiteaAdminSecretObject() corev1.Secret {
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GiteaAdminSecret,
			Namespace: GiteaNamespace,
		},
	}
}

func GiteaBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	return fmt.Sprintf(GiteaIngressURL, config.Protocol, config.Port)
}
