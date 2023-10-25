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

	// Create cluster
	cluster, err := NewCluster("testcase", "v1.26.3", "", "", "")
	if err != nil {
		t.Fatalf("Initializing cluster resource: %v", err)
	}

	// Create registry
	err = cluster.ReconcileRegistry(ctx)
	if err != nil {
		t.Fatalf("Error reconciling registry: %v", err)
	}

	// Get resulting container
	container, err := cluster.getRegistryContainer(ctx, dockerCli)
	if err != nil {
		t.Fatalf("Error getting registry container after reconcile: %v", err)
	}
	if container == nil {
		t.Fatal("Expected registry container after reconcile but got nil")
	}

	// Run a second reconcile to validate idempotency
	err = cluster.ReconcileRegistry(ctx)
	if err != nil {
		t.Fatalf("Error reconciling registry: %v", err)
	}

	// Cleanup
	if err = dockerCli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
		t.Fatalf("Error removing registry docker container after reconcile: %v", err)
	}
}
