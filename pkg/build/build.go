package build

import (
	"context"
	"fmt"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/controllers"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type Build struct {
	name              string
	kindConfigPath    string
	kubeConfigPath    string
	kubeVersion       string
	extraPortsMapping string
	customPackageDirs []string
	exitOnSync        bool
	scheme            *runtime.Scheme
	CancelFunc        context.CancelFunc
}

func NewBuild(name, kubeVersion, kubeConfigPath, kindConfigPath, extraPortsMapping string, customPackageDirs []string, exitOnSync bool, scheme *runtime.Scheme, ctxCancel context.CancelFunc) *Build {
	return &Build{
		name:              name,
		kindConfigPath:    kindConfigPath,
		kubeConfigPath:    kubeConfigPath,
		kubeVersion:       kubeVersion,
		extraPortsMapping: extraPortsMapping,
		customPackageDirs: customPackageDirs,
		exitOnSync:        exitOnSync,
		scheme:            scheme,
		CancelFunc:        ctxCancel,
	}
}

func (b *Build) ReconcileKindCluster(ctx context.Context, recreateCluster bool) error {
	// Initialize Kind Cluster
	cluster, err := kind.NewCluster(b.name, b.kubeVersion, b.kubeConfigPath, b.kindConfigPath, b.extraPortsMapping)
	if err != nil {
		setupLog.Error(err, "Error Creating kind cluster")
		return err
	}

	// Build Kind cluster
	if err := cluster.Reconcile(ctx, recreateCluster); err != nil {
		setupLog.Error(err, "Error starting kind cluster")
		return err
	}

	// Create Kube Config for Kind cluster
	if err := cluster.ExportKubeConfig(b.name, false); err != nil {
		setupLog.Error(err, "Error exporting kubeconfig from kind cluster")
		return err
	}
	return nil
}

func (b *Build) GetKubeConfig() (*rest.Config, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", b.kubeConfigPath)
	if err != nil {
		setupLog.Error(err, "Error building kubeconfig from kind cluster")
		return nil, err
	}
	return kubeConfig, nil
}

func (b *Build) GetKubeClient(kubeConfig *rest.Config) (client.Client, error) {
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: b.scheme})
	if err != nil {
		setupLog.Error(err, "Error creating kubernetes client")
		return nil, err
	}
	return kubeClient, nil
}

func (b *Build) ReconcileCRDs(ctx context.Context, kubeClient client.Client) error {
	// Ensure idpbuilder CRDs
	if err := controllers.EnsureCRDs(ctx, b.scheme, kubeClient); err != nil {
		setupLog.Error(err, "Error creating idpbuilder CRDs")
		return err
	}
	return nil
}

func (b *Build) RunControllers(ctx context.Context, mgr manager.Manager, exitCh chan error) error {
	return controllers.RunControllers(ctx, mgr, exitCh, b.CancelFunc, b.exitOnSync)
}

func (b *Build) Run(ctx context.Context, recreateCluster bool) error {
	managerExit := make(chan error)

	setupLog.Info("Creating kind cluster")
	if err := b.ReconcileKindCluster(ctx, recreateCluster); err != nil {
		return err
	}

	setupLog.Info("Getting Kube config")
	kubeConfig, err := b.GetKubeConfig()
	if err != nil {
		return err
	}

	setupLog.Info("Getting Kube client")
	kubeClient, err := b.GetKubeClient(kubeConfig)
	if err != nil {
		return err
	}

	setupLog.Info("Adding CRDs to the cluster")
	if err := b.ReconcileCRDs(ctx, kubeClient); err != nil {
		return err
	}

	setupLog.Info("Creating controller manager")
	// Create controller manager
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: b.scheme,
	})
	if err != nil {
		setupLog.Error(err, "Error creating controller manager")
		return err
	}

	setupLog.Info("Running controllers")
	if err := b.RunControllers(ctx, mgr, managerExit); err != nil {
		setupLog.Error(err, "Error running controllers")
		return err
	}

	localBuild := v1alpha1.Localbuild{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.name,
		},
	}

	cliStartTime := time.Now().Format(time.RFC3339Nano)

	setupLog.Info("Creating localbuild resource")
	_, err = controllerutil.CreateOrUpdate(ctx, kubeClient, &localBuild, func() error {
		if localBuild.ObjectMeta.Annotations == nil {
			localBuild.ObjectMeta.Annotations = map[string]string{}
		}
		localBuild.ObjectMeta.Annotations[v1alpha1.CliStartTimeAnnotation] = cliStartTime
		localBuild.Spec = v1alpha1.LocalbuildSpec{
			PackageConfigs: v1alpha1.PackageConfigsSpec{
				Argo: v1alpha1.ArgoPackageConfigSpec{
					Enabled: true,
				},
				EmbeddedArgoApplications: v1alpha1.EmbeddedArgoApplicationsPackageConfigSpec{
					Enabled: true,
				},
				CustomPackageDirs: b.customPackageDirs,
			},
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("creating localbuild resource: %w", err)
	}

	if err != nil {
		setupLog.Error(err, "Error creating localbuild resource")
		return err
	}

	err = <-managerExit
	close(managerExit)
	return err
}
