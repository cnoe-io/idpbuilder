package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
