package custompackage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestReconcileCustomPkg(t *testing.T) {
	s := k8sruntime.NewScheme()
	sb := k8sruntime.NewSchemeBuilder(
		v1.AddToScheme,
		argov1alpha1.AddToScheme,
		v1alpha1.AddToScheme,
	)
	sb.AddToScheme(s)
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "resources"),
			"../localbuild/resources/argo/install.yaml",
		},
		ErrorIfCRDPathMissing: true,
		Scheme:                s,
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "bin", "k8s",
			fmt.Sprintf("1.27.1-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("Starting testenv: %v", err)
	}
	defer testEnv.Stop()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: s,
	})
	if err != nil {
		t.Fatalf("getting manager: %v", err)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	stoppedCh := make(chan error)
	go func() {
		err := mgr.Start(ctx)
		stoppedCh <- err
	}()

	defer func() {
		ctxCancel()
		err := <-stoppedCh
		if err != nil {
			t.Errorf("Starting controller manager: %v", err)
			t.FailNow()
		}
	}()

	r := &Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("test-custompkg-controller"),
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd %v", err)
	}
	customPkgs := []v1alpha1.CustomPackage{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "test",
				UID:       "abc",
			},
			Spec: v1alpha1.CustomPackageSpec{
				Replicate:           true,
				GitServerURL:        "https://cnoe.io",
				InternalGitServeURL: "http://internal.cnoe.io",
				ArgoCD: v1alpha1.ArgoCDPackageSpec{
					ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir/app.yaml"),
					Name:            "my-app",
					Namespace:       "argocd",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test2",
				Namespace: "test",
				UID:       "abc",
			},
			Spec: v1alpha1.CustomPackageSpec{
				Replicate:           false,
				GitServerURL:        "https://cnoe.io",
				InternalGitServeURL: "http://cnoe.io/internal",
				ArgoCD: v1alpha1.ArgoCDPackageSpec{
					ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir2/exampleApp.yaml"),
					Name:            "guestbook",
					Namespace:       "argocd",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test3",
				Namespace: "test",
				UID:       "abc",
			},
			Spec: v1alpha1.CustomPackageSpec{
				Replicate:           true,
				GitServerURL:        "https://cnoe.io",
				InternalGitServeURL: "http://internal.cnoe.io",
				ArgoCD: v1alpha1.ArgoCDPackageSpec{
					ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir/app2.yaml"),
					Name:            "my-app2",
					Namespace:       "argocd",
				},
			},
		},
	}

	for _, n := range []string{"argocd", "test"} {
		ns := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: n,
			},
		}
		err = mgr.GetClient().Create(context.Background(), &ns)
		if err != nil {
			t.Fatalf("creating test ns: %v", err)
		}
	}

	for i := range customPkgs {
		_, err = r.reconcileCustomPackage(context.Background(), &customPkgs[i])
		if err != nil {
			t.Fatalf("reconciling custom packages %v", err)
		}
	}
	time.Sleep(1 * time.Second)
	// verify repo.
	c := mgr.GetClient()
	repo := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoName("my-app", "test/resources/customPackages/testDir/busybox"),
			Namespace: "test",
		},
	}
	err = c.Get(context.Background(), client.ObjectKeyFromObject(&repo), &repo)
	if err != nil {
		t.Fatalf("getting my-app-busybox git repo %v", err)
	}

	p, _ := filepath.Abs("test/resources/customPackages/testDir/busybox")
	expectedRepo := v1alpha1.GitRepository{
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Type: "local",
				Path: p,
			},
			GitURL:         "https://cnoe.io",
			InternalGitURL: "http://internal.cnoe.io",
		},
	}
	ok := reflect.DeepEqual(repo.Spec, expectedRepo.Spec)
	if !ok {
		t.Fatalf("expected spec does not match")
	}

	// verify argocd apps
	localApp := argov1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app",
			Namespace: "argocd",
		},
	}
	err = c.Get(context.Background(), client.ObjectKeyFromObject(&localApp), &localApp)
	if err != nil {
		t.Fatalf("failed getting my-app %v", err)
	}
	if strings.HasPrefix(localApp.Spec.Source.RepoURL, "cnoe://") {
		t.Fatalf("cnoe:// prefix should be removed")
	}

	for _, n := range []string{"guestbook", "guestbook2"} {
		err = c.Get(context.Background(), client.ObjectKeyFromObject(&localApp), &localApp)
		if err != nil {
			t.Fatalf("expected %s arogapp : %v", n, err)
		}
	}

	localApp2 := argov1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app2",
			Namespace: "argocd",
		},
	}
	err = c.Get(context.Background(), client.ObjectKeyFromObject(&localApp2), &localApp2)
	if err != nil {
		t.Fatalf("failed getting my-app2 %v", err)
	}

	if strings.HasPrefix(localApp2.Spec.Sources[0].RepoURL, "cnoe://") {
		t.Fatalf("cnoe:// prefix should be removed")
	}
}
