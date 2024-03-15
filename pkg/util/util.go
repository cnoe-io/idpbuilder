package util

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetCLIStartTimeAnnotationValue(annotations map[string]string) (string, error) {
	if annotations == nil {
		return "", fmt.Errorf("this object's annotation is nil")
	}
	timeStamp, ok := annotations[v1alpha1.CliStartTimeAnnotation]
	if ok {
		return timeStamp, nil
	}
	return "", fmt.Errorf("expected annotation, %s, not found", v1alpha1.CliStartTimeAnnotation)
}

func SetCLIStartTimeAnnotationValue(annotations map[string]string, timeStamp string) {
	if timeStamp != "" && annotations != nil {
		annotations[v1alpha1.CliStartTimeAnnotation] = timeStamp
	}
}

func SetLastObservedSyncTimeAnnotationValue(annotations map[string]string, timeStamp string) {
	if timeStamp != "" && annotations != nil {
		annotations[v1alpha1.LastObservedCLIStartTimeAnnotation] = timeStamp
	}
}

func GetLastObservedSyncTimeAnnotationValue(annotations map[string]string) (string, error) {
	if annotations == nil {
		return "", fmt.Errorf("this object's annotation is nil")
	}
	timeStamp, ok := annotations[v1alpha1.LastObservedCLIStartTimeAnnotation]
	if ok {
		return timeStamp, nil
	}
	return "", fmt.Errorf("expected annotation, %s, not found", v1alpha1.LastObservedCLIStartTimeAnnotation)
}

func UpdateSyncAnnotation(ctx context.Context, kubeClient client.Client, obj client.Object) error {
	timeStamp, err := GetCLIStartTimeAnnotationValue(obj.GetAnnotations())
	if err != nil {
		return err
	}
	annotations := make(map[string]string, 1)
	SetLastObservedSyncTimeAnnotationValue(annotations, timeStamp)
	// MUST be unstructured to avoid managing fields we do not care about.
	u := unstructured.Unstructured{}
	u.SetAnnotations(annotations)
	u.SetName(obj.GetName())
	u.SetNamespace(obj.GetNamespace())
	u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	return kubeClient.Patch(ctx, &u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}
