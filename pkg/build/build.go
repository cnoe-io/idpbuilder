package build

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/controllers"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type Build struct {
	name                 string
	cfg                  util.CorePackageTemplateConfig
	kindConfigPath       string
	kubeConfigPath       string
	kubeVersion          string
	extraPortsMapping    string
	customPackageDirs    []string
	customPackageUrls    []string
	packageCustomization map[string]v1alpha1.PackageCustomization
	exitOnSync           bool
	scheme               *runtime.Scheme
	CancelFunc           context.CancelFunc
}

type NewBuildOptions struct {
	Name                 string
	TemplateData         util.CorePackageTemplateConfig
	KindConfigPath       string
	KubeConfigPath       string
	KubeVersion          string
	ExtraPortsMapping    string
	CustomPackageDirs    []string
	CustomPackageUrls    []string
	PackageCustomization map[string]v1alpha1.PackageCustomization
	ExitOnSync           bool
	Scheme               *runtime.Scheme
	CancelFunc           context.CancelFunc
}

func NewBuild(opts NewBuildOptions) *Build {
	return &Build{
		name:                 opts.Name,
		kindConfigPath:       opts.KindConfigPath,
		kubeConfigPath:       opts.KubeConfigPath,
		kubeVersion:          opts.KubeVersion,
		extraPortsMapping:    opts.ExtraPortsMapping,
		customPackageDirs:    opts.CustomPackageDirs,
		customPackageUrls:    opts.CustomPackageUrls,
		packageCustomization: opts.PackageCustomization,
		exitOnSync:           opts.ExitOnSync,
		scheme:               opts.Scheme,
		cfg:                  opts.TemplateData,
		CancelFunc:           opts.CancelFunc,
	}
}

func (b *Build) ReconcileKindCluster(ctx context.Context, recreateCluster bool) error {
	// Initialize Kind Cluster
	cluster, err := kind.NewCluster(b.name, b.kubeVersion, b.kubeConfigPath, b.kindConfigPath, b.extraPortsMapping, b.cfg)
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
	if err := controllers.EnsureCRDs(ctx, b.scheme, kubeClient, b.cfg); err != nil {
		setupLog.Error(err, "Error creating idpbuilder CRDs")
		return err
	}
	return nil
}

func (b *Build) RunControllers(ctx context.Context, mgr manager.Manager, exitCh chan error, tmpDir string) error {
	return controllers.RunControllers(ctx, mgr, exitCh, b.CancelFunc, b.exitOnSync, b.cfg, tmpDir)
}

func (b *Build) Run(ctx context.Context, recreateCluster bool) error {
	managerExit := make(chan error)

	setupLog.Info("Creating kind cluster")
	if err := b.ReconcileKindCluster(ctx, recreateCluster); err != nil {
		return err
	}

	setupLog.V(1).Info("Getting Kube config")
	kubeConfig, err := b.GetKubeConfig()
	if err != nil {
		return err
	}

	setupLog.V(1).Info("Getting Kube client")
	kubeClient, err := b.GetKubeClient(kubeConfig)
	if err != nil {
		return err
	}

	setupLog.Info("Adding CRDs to the cluster")
	if err := b.ReconcileCRDs(ctx, kubeClient); err != nil {
		return err
	}

	setupLog.V(1).Info("Creating controller manager")
	// Create controller manager
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: b.scheme,
		Metrics: server.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		setupLog.Error(err, "Error creating controller manager")
		return err
	}

	dir, err := os.MkdirTemp("", fmt.Sprintf("%s-%s-", globals.ProjectName, b.name))
	if err != nil {
		setupLog.Error(err, "creating temp dir")
		return err
	}
	defer os.RemoveAll(dir)
	setupLog.V(1).Info("Created temp directory for cloning repositories", "dir", dir)

	setupLog.V(1).Info("Running controllers")
	if err := b.RunControllers(ctx, mgr, managerExit, dir); err != nil {
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
					Enabled:              true,
					PackageCustomization: b.packageCustomization,
				},
				CustomPackageDirs: b.customPackageDirs,
				CustomPackageUrls: b.customPackageUrls,
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
