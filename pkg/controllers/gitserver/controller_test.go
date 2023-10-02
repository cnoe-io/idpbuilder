package gitserver

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	ResourceName      = "test"
	ResourceNamespace = "default"

	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

func TestGitServerController(t *testing.T) {
	//specify testEnv configuration
	scheme := k8s.GetScheme()
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "resources")},
		ErrorIfCRDPathMissing: true,
		Scheme:                scheme,
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "bin", "k8s",
			fmt.Sprintf("1.27.1-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	//start testEnv
	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("Starting testenv: %v", err)
	}
	defer testEnv.Stop()

	// Start controller
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
	})
	k8sClient := k8sManager.GetClient()
	if err != nil {
		t.Fatalf("Creating controller manager: %v", err)
	}
	if err := (&GitServerReconciler{
		Client: k8sClient,
		Scheme: k8sManager.GetScheme(),
		Content: fstest.MapFS{
			"Dockerfile": &fstest.MapFile{
				Data: []byte("FROM nginx\n"),
				Mode: 0666,
			},
		},
	}).SetupWithManager(k8sManager); err != nil {
		t.Fatalf("Unable to create controller with manager: %v", err)
	}

	// Run manager in background
	ctx, ctxCancel := context.WithCancel(context.Background())
	stoppedCh := make(chan error)
	go func() {
		err := k8sManager.Start(ctx)
		t.Log("Controller stopped")
		stoppedCh <- err
	}()

	// Defer controller shutdown
	defer func() {
		ctxCancel()
		err := <-stoppedCh
		if err != nil {
			t.Errorf("Starting controller manager: %v", err)
			t.FailNow()
		}
	}()

	// Create GitServer resource
	resource := &v1alpha1.GitServer{
		TypeMeta: v1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.Identifier(),
			Kind:       "GitServer",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      ResourceName,
			Namespace: ResourceNamespace,
			UID:       "test-uid",
		},
	}
	if err := k8sClient.Create(ctx, resource); err != nil {
		t.Fatalf("Creating resource: %v", err)
	}

	// Wait for GitServer to become available
	endTime := time.Now().Add(timeout)
	for {
		if time.Now().After(endTime) {
			t.Fatal("Timed out waiting for resource available")
		}

		var gotResource v1alpha1.GitServer
		if err = k8sClient.Get(ctx, client.ObjectKey{
			Name:      ResourceName,
			Namespace: ResourceNamespace,
		}, &gotResource); err != nil {
			t.Logf("Failed getting resource (might be ok though): %v", err)
			continue
		}

		if gotResource.Status.DeploymentAvailable {
			break
		}

		time.Sleep(interval)
	}
}
