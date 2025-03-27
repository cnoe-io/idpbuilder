package util

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetConfig(ctx context.Context) (v1alpha1.BuildCustomizationSpec, error) {
	b := v1alpha1.BuildCustomizationSpec{}

	kubeConfig, err := GetKubeConfig()
	if err != nil {
		return b, fmt.Errorf("getting kube config: %w", err)
	}

	kubeClient, err := GetKubeClient(kubeConfig)
	if err != nil {
		return b, fmt.Errorf("getting kube client: %w", err)
	}

	list, err := getLocalBuild(ctx, kubeClient)
	if err != nil {
		return b, err
	}

	// TODO: We assume that only one LocalBuild exists !
	return list.Items[0].Spec.BuildCustomization, nil
}

func getLocalBuild(ctx context.Context, kubeClient client.Client) (v1alpha1.LocalbuildList, error) {
	localBuildList := v1alpha1.LocalbuildList{}
	return localBuildList, kubeClient.List(ctx, &localBuildList)
}
