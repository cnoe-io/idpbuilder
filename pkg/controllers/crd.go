package controllers

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed resources/*.yaml
var crdFS embed.FS

func getK8sResources(scheme *runtime.Scheme, templateData any) ([]client.Object, error) {
	rawResources, err := util.ConvertFSToBytes(crdFS, "resources", templateData)
	if err != nil {
		return nil, err
	}

	return k8s.ConvertRawResourcesToObjects(scheme, rawResources)
}

func EnsureCRD(ctx context.Context, scheme *runtime.Scheme, kubeClient client.Client, obj client.Object) error {
	logger := log.FromContext(ctx)

	// Check if the CRD already exists
	crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
	if !ok {
		return fmt.Errorf("non crd object passed to EnsureCRD: %v", obj)
	}
	var curCRD apiextensionsv1.CustomResourceDefinition
	err := kubeClient.Get(
		ctx,
		types.NamespacedName{Name: obj.GetName(), Namespace: "default"},
		&curCRD)

	switch {
	case apierrors.IsNotFound(err):
		if err := kubeClient.Create(ctx, obj); err != nil {
			logger.Error(err, "Unable to create CRD", "resource", obj)
			return err
		}
	case err != nil:
		logger.Error(err, "Unable to get CRD during initial check", "resource", obj)
		return err
	default:
		crd.SetResourceVersion(curCRD.GetResourceVersion())
		if err = kubeClient.Update(ctx, crd); err != nil {
			logger.Error(err, "Updating CRD", "resource", obj)
			return err
		}
	}

	// There is some async work before the CRD actually exists, wait for this
	for {
		if err := kubeClient.Get(
			ctx,
			types.NamespacedName{Name: obj.GetName(), Namespace: "default"},
			&curCRD,
		); err != nil {
			logger.Error(err, "Failed to get CRD", "crd name", obj.GetName())
			return err
		}
		crdEstablished := false
		for _, cond := range curCRD.Status.Conditions {
			if cond.Type == apiextensionsv1.Established {
				if cond.Status == apiextensionsv1.ConditionTrue {
					crdEstablished = true
				}
			}
		}
		if crdEstablished {
			break
		} else {
			logger.V(1).Info("crd not yet established, waiting.", "crd name", obj.GetName())
		}
		time.Sleep(time.Duration(time.Duration.Milliseconds(500)))
	}
	return nil
}

func EnsureCRDs(ctx context.Context, scheme *runtime.Scheme, kubeClient client.Client, templateData any) error {
	installObjs, err := getK8sResources(scheme, templateData)
	if err != nil {
		return err
	}

	for _, obj := range installObjs {
		if err = EnsureCRD(ctx, scheme, kubeClient, obj); err != nil {
			return err
		}
	}

	return nil
}
