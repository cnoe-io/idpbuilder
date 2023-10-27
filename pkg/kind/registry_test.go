package kind

import (
	"context"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/docker/docker/api/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestReconcileRegistry(t *testing.T) {
	ctrl.Log.WithName("test")
	ctx := context.Background()

	dockerCli, err := docker.GetDockerClient()
	if err != nil {
		t.Fatalf("Error getting docker client: %v", err)
	}
	defer dockerCli.Close()

	kindNetwork, err := dockerCli.NetworkCreate(ctx, "kind", types.NetworkCreate{})
	if err != nil {
		t.Fatalf("Failed creaking kind network: %v", err)
	}
	defer dockerCli.NetworkRemove(ctx, kindNetwork.ID)

	// Create cluster
	cluster, err := NewCluster("testcase", "v1.26.3", "", "", "")
	if err != nil {
		t.Fatalf("Initializing cluster resource: %v", err)
	}

	// Create registry
	defer dockerCli.ContainerRemove(ctx, cluster.getRegistryContainerName(), types.ContainerRemoveOptions{Force: true})
	waitTimeout := time.Second * 90
	waitInterval := time.Second * 3
	endTime := time.Now().Add(waitTimeout)

	for {
		if time.Now().After(endTime) {
			t.Fatalf("Timed out waiting for registry. recent error: %v", err)
		}
		err = cluster.ReconcileRegistry(ctx)
		if err == nil {
			break
		}
		t.Logf("Failed to reconcile: %v", err)
		dockerCli.ContainerRemove(ctx, cluster.getRegistryContainerName(), types.ContainerRemoveOptions{Force: true})
		time.Sleep(waitInterval)
	}

	// Get resulting container
	container, err := cluster.getRegistryContainer(ctx, dockerCli)
	if err != nil {
		t.Fatalf("Error getting registry container after reconcile: %v", err)
	}
	defer dockerCli.ImageRemove(ctx, container.ImageID, types.ImageRemoveOptions{})

	if container == nil {
		t.Fatal("Expected registry container after reconcile but got nil")
	}

	// Run a second reconcile to validate idempotency
	err = cluster.ReconcileRegistry(ctx)
	if err != nil {
		t.Fatalf("Error reconciling registry: %v", err)
	}
}

