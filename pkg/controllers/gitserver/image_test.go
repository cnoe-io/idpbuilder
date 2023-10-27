package gitserver

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/apps"
	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestReconcileGitServerImage(t *testing.T) {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	appsFS, err := apps.GetAppsFS()
	if err != nil {
		t.Fatalf("Getting apps FS: %v", err)
	}

	ctx := context.Background()
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

	dockerClient, err := docker.GetDockerClient()
	if err != nil {
		t.Errorf("Getting docker client: %v", err)
	}
	defer dockerClient.Close()
	reader, err := dockerClient.ImagePull(ctx, "docker.io/library/registry:2", types.ImagePullOptions{})
	defer reader.Close()
	// blocks until pull is completed
	io.Copy(os.Stdout, reader)
	if err != nil {
		t.Fatalf("failed pulilng registry image: %v", err)
	}

	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "docker.io/library/registry:2",
		Tty:   false,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", kind.InternalRegistryPort)): struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", kind.InternalRegistryPort)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", kind.ExposedRegistryPort),
				},
			},
		},
	}, nil, nil, "testcase-registry")
	if err != nil {
		t.Fatalf("failed creating registry container %v", err)
	}

	defer dockerClient.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})

	waitTimeout := time.Second * 90
	waitInterval := time.Second * 3
	endTime := time.Now().Add(waitTimeout)
	// avoid failures with the container that uses port 5001 in TestReconcileRegistry
	for {
		if time.Now().After(endTime) {
			t.Fatalf("Timed out waiting for idpbuilder-testcase-registry container to terminate")
		}
		containers, _ := dockerClient.ContainerList(ctx, types.ContainerListOptions{
			Filters: filters.NewArgs(filters.KeyValuePair{
				Key:   "name",
				Value: "idpbuilder-testcase-registry",
			}),
		})
		if len(containers) == 0 {
			break
		}
		time.Sleep(waitInterval)
	}
	err = dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		t.Fatalf("failed starting container %v", err)
	}

	_, err = r.reconcileGitServerImage(ctx, controllerruntime.Request{}, &resource)
	if err != nil {
		t.Fatalf("reconcile error: %v", err)
	}

	if !strings.HasPrefix(resource.Status.ImageID, "sha256") {
		t.Fatalf("Invalid or no Image ID in status: %q", resource.Status.ImageID)
	}
	imageNameID := fmt.Sprintf("%s@%s", GetImageTag(&resource), resource.Status.ImageID)
	_, err = dockerClient.ImageRemove(ctx, imageNameID, types.ImageRemoveOptions{})
	if err != nil {
		t.Errorf("Removing docker image: %v", err)
	}
}
