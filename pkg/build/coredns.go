package build

import (
	"context"
	"embed"
	"fmt"

	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	coreDNSTemplatePath = "templates/coredns"
)

//go:embed templates
var templates embed.FS

func setupCoreDNS(ctx context.Context, kubeClient client.Client, scheme *runtime.Scheme, templateData util.CorePackageTemplateConfig) error {
	checkCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns-conf-default",
			Namespace: "kube-system",
		},
	}
	err := kubeClient.Get(ctx, client.ObjectKeyFromObject(checkCM), checkCM)
	if err == nil {
		return nil
	}

	objs, err := k8s.BuildCustomizedObjects("", coreDNSTemplatePath, templates, scheme, templateData)
	if err != nil {
		return fmt.Errorf("rendering embedded coredns files: %w", err)
	}

	for i := range objs {
		obj := objs[i]
		switch t := obj.(type) {
		case *appsv1.Deployment:
			dep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      t.Name,
					Namespace: t.Namespace,
				},
			}
			_, err = controllerutil.CreateOrUpdate(ctx, kubeClient, dep, func() error {
				dep.Spec = t.Spec
				return nil
			})
			if err != nil {
				return fmt.Errorf("creating/updating deployment: %w", err)
			}
		case *corev1.ConfigMap:
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      t.Name,
					Namespace: t.Namespace,
				},
			}
			_, err = controllerutil.CreateOrUpdate(ctx, kubeClient, cm, func() error {
				cm.Data = t.Data
				return nil
			})
			if err != nil {
				return fmt.Errorf("creating/updating configmap: %w", err)
			}
		}
	}
	return nil
}
