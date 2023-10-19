package gitserver

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/apps"
	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"testing"
)

func TestReconcileGitServerImage(t *testing.T) {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "registry:2", // registry image does expose port 5000
		WaitingFor:   wait.ForListeningPort("5000/tcp"),
		ExposedPorts: []string{"5001:5000/tcp"},
	}
	registryC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := registryC.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	appsFS, err := apps.GetAppsFS()
	if err != nil {
		t.Fatalf("Getting apps FS: %v", err)
	}

	r := GitServerReconciler{
		Content: appsFS,
	}
	resource := v1alpha1.GitServer{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testcase",
		},
		Spec: v1alpha1.GitServerSpec{
			Source: v1alpha1.GitServerSource{
				Embedded: true,
			},
		},
	}

	_, err = r.reconcileGitServerImage(ctx, controllerruntime.Request{}, &resource)
	if err != nil {
		t.Errorf("reconcile error: %v", err)
	}

	if !strings.HasPrefix(resource.Status.ImageID, "sha256") {
		t.Errorf("Invalid or no Image ID in status: %q", resource.Status.ImageID)
	}

	dockerClient, err := docker.GetDockerClient()
	if err != nil {
		t.Errorf("Getting docker client: %v", err)
	}

	imageNameID := fmt.Sprintf("%s@%s", GetImageTag(&resource), resource.Status.ImageID)
	_, err = dockerClient.ImageRemove(ctx, imageNameID, types.ImageRemoveOptions{})
	if err != nil {
		t.Errorf("Removing docker image: %v", err)
	}

}
