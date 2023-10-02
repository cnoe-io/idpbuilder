package gitserver

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/apps"
	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetImageTag(resource *v1alpha1.GitServer) string {
	return fmt.Sprintf("localhost:%d/%s-%s-%s-gitserver", kind.ExposedRegistryPort, globals.ProjectName, resource.Namespace, resource.Name)
}

func GetImageUrl(resource *v1alpha1.GitServer) string {
	if resource.Spec.Source.Embedded {
		return fmt.Sprintf("%s@%s", GetImageTag(resource), resource.Status.ImageID)
	}
	return resource.Spec.Source.Image
}

func (g *GitServerReconciler) reconcileGitServerImage(ctx context.Context, req ctrl.Request, resource *v1alpha1.GitServer) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// If we're not using the embedded source, bail
	if !resource.Spec.Source.Embedded {
		log.Info("Not using embedded source, skipping image building")
		return ctrl.Result{}, nil
	}

	dockerClient, err := docker.GetDockerClient()
	if err != nil {
		return ctrl.Result{}, err
	}
	defer dockerClient.Close()

	imageTag := GetImageTag(resource)

	// Build image
	_, err = apps.BuildAppsImage(ctx, dockerClient, []string{imageTag}, map[string]string{
		gitServerLabelKey: resource.GetName(),
	}, g.Content)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Push image
	regImgId, err := apps.PushImage(ctx, dockerClient, imageTag)
	if err != nil {
		return ctrl.Result{}, err
	}

	if regImgId == nil {
		return ctrl.Result{}, fmt.Errorf("failed to get registry image id after push")
	}

	resource.Status.ImageID = *regImgId

	return ctrl.Result{}, nil
}
