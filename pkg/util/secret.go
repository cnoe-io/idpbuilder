package util

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSecretByName(ctx context.Context, kubeClient client.Client, ns, name string) (v1.Secret, error) {
	s := v1.Secret{}
	return s, kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, &s)
}
