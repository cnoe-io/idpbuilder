package k8s

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetKubeClient() (client.Client, error) {
	conf, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return nil, err
	}
	return client.New(conf, client.Options{Scheme: GetScheme()})
}

func EnsureObject(ctx context.Context, kubeClient client.Client, obj client.Object, namespace string) error {
	curObj := &unstructured.Unstructured{}
	curObj.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	// Fallback to object's namespace
	if namespace == "" {
		namespace = obj.GetNamespace()
	}

	// Get Object if it exists
	err := kubeClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: namespace,
			Name:      obj.GetName(),
		},
		curObj,
	)

	if err == nil {
		// Object already exists
		return nil
	}

	err = kubeClient.Create(ctx, obj)
	if err != nil {
		return err
	}

	// hacky way to restore the GVK for the object after create corrupts it. didn't dig. not sure why?
	obj.GetObjectKind().SetGroupVersionKind(curObj.GroupVersionKind())
	return nil
}

func EnsureNamespace(ctx context.Context, kubeClient client.Client, name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	err := kubeClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return kubeClient.Create(ctx, ns)
		} else {
			return fmt.Errorf("getting namespace %s: %w", name, err)
		}
	}
	return nil
}
