package localbuild

import (
	"context"
	"fmt"
	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"path/filepath"
	"reflect"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"strings"
	"testing"
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
			"resources/argo/install.yaml",
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

	r := &LocalbuildReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		CancelFunc:     nil,
		shouldShutdown: false,
	}
	customPkgs := []v1alpha1.CustomPackageSpec{
		{
			Directory: "test/resources/customPackages/testDir",
		},
		{
			Directory: "test/resources/customPackages/testDir2",
		},
	}

	res := v1alpha1.Localbuild{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			UID:  "uid",
		},
		Spec: v1alpha1.LocalbuildSpec{
			PackageConfigs: v1alpha1.PackageConfigsSpec{
				GitConfig:                v1alpha1.GitConfigSpec{},
				Argo:                     v1alpha1.ArgoPackageConfigSpec{},
				EmbeddedArgoApplications: v1alpha1.EmbeddedArgoApplicationsPackageConfigSpec{},
				CustomPackages:           customPkgs,
			},
		},
		Status: v1alpha1.LocalbuildStatus{
			Gitea: v1alpha1.GiteaStatus{
				Available:                true,
				ExternalURL:              "https://cnoe.io",
				InternalURL:              "http://internal.cnoe.io",
				AdminUserSecretName:      "abc",
				AdminUserSecretNamespace: "abc",
			},
		},
	}

	for _, n := range []string{"argocd", globals.GetProjectNamespace(res.Name)} {
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
		_, err = r.reconcileCustomPkg(context.Background(), &res, customPkgs[i])
		if err != nil {
			t.Fatalf("reconciling custom packages %v", err)
		}
	}

	// verify repo.
	c := mgr.GetClient()
	repo := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app-busybox",
			Namespace: globals.GetProjectNamespace(res.Name),
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
			GitURL: "https://cnoe.io",
			SecretRef: v1alpha1.SecretReference{
				Name:      "abc",
				Namespace: "abc",
			},
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
}
