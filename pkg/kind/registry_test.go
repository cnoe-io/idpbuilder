package kind

import (
	"context"
	"testing"

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
	err = cluster.ReconcileRegistry(ctx)
	defer dockerCli.ContainerRemove(ctx, cluster.getRegistryContainerName(), types.ContainerRemoveOptions{Force: true})
	if err != nil {
		t.Fatalf("Error reconciling registry: %v", err)
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
