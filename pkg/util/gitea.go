package util

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func PatchPasswordSecret(ctx context.Context, kubeClient client.Client, ns string, secretName string, pass string) error {
	sec, err := GetSecretByName(ctx, kubeClient, ns, secretName)
	if err != nil {
		return fmt.Errorf("getting secret to patch fails: %w", err)
	}
	u := unstructured.Unstructured{}
	u.SetName(sec.GetName())
	u.SetNamespace(sec.GetNamespace())
	u.SetGroupVersionKind(sec.GetObjectKind().GroupVersionKind())

	err = unstructured.SetNestedField(u.Object, base64.StdEncoding.EncodeToString([]byte(pass)), "data", "password")
	if err != nil {
		return fmt.Errorf("setting password field: %w", err)
	}

	return kubeClient.Patch(ctx, &u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}

func GiteaBaseUrl(config v1alpha1.BuildCustomizationSpec) string {
	return fmt.Sprintf(GiteaIngressURL, config.Protocol, config.Port)
}
