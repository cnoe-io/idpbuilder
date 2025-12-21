package build

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/controllers"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	cfg                  v1alpha1.BuildCustomizationSpec
	kindConfigPath       string
	kubeConfigPath       string
	kubeVersion          string
	extraPortsMapping    string
	registryConfig       []string
	customPackageFiles   []string
	customPackageDirs    []string
	customPackageUrls    []string
	packageCustomization map[string]v1alpha1.PackageCustomization
	exitOnSync           bool
	scheme               *runtime.Scheme
	CancelFunc           context.CancelFunc
}

type NewBuildOptions struct {
	Name                 string
	TemplateData         v1alpha1.BuildCustomizationSpec
	KindConfigPath       string
	KubeConfigPath       string
	KubeVersion          string
	ExtraPortsMapping    string
	RegistryConfig       []string
	CustomPackageFiles   []string
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
		registryConfig:       opts.RegistryConfig,
		customPackageFiles:   opts.CustomPackageFiles,
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
	cluster, err := kind.NewCluster(b.name, b.kubeVersion, b.kubeConfigPath, b.kindConfigPath, b.extraPortsMapping, b.registryConfig, b.cfg, setupLog)
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

func (b *Build) isCompatible(ctx context.Context, kubeClient client.Client) (bool, error) {
	localBuild := v1alpha1.Localbuild{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.name,
		},
	}

	err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&localBuild), &localBuild)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	ok := isBuildCustomizationSpecEqual(b.cfg, localBuild.Spec.BuildCustomization)

	if ok {
		return ok, nil
	}

	existing, given := localBuild.Spec.BuildCustomization, b.cfg
	existing.SelfSignedCert = ""
	given.SelfSignedCert = ""

	return false, fmt.Errorf("provided command flags and existing configurations are incompatible. please recreate the cluster. "+
		"existing: %+v, given: %+v",
		existing, given)
}

func (b *Build) Run(ctx context.Context, recreateCluster bool) error {
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

	setupLog.Info("Setting up CoreDNS")
	err = setupCoreDNS(ctx, kubeClient, b.scheme, b.cfg)
	if err != nil {
		return err
	}

	setupLog.Info("Setting up TLS certificate")
	cert, err := setupSelfSignedCertificate(ctx, setupLog, kubeClient, b.cfg)
	if err != nil {
		return err
	}
	b.cfg.SelfSignedCert = string(cert)

	setupLog.V(1).Info("Checking for incompatible options from a previous run")
	ok, err := b.isCompatible(ctx, kubeClient)
	if err != nil {
		setupLog.Error(err, "Error while checking incompatible flags")
		return err
	}
	if !ok {
		return err
	}

	managerExit := make(chan error)

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
			BuildCustomization: b.cfg,
			PackageConfigs: v1alpha1.PackageConfigsSpec{
				Argo: v1alpha1.ArgoPackageConfigSpec{
					Enabled: true,
				},
				EmbeddedArgoApplications: v1alpha1.EmbeddedArgoApplicationsPackageConfigSpec{
					Enabled: true,
				},
				CustomPackageDirs:        b.customPackageDirs,
				CustomPackageFiles:       b.customPackageFiles,
				CustomPackageUrls:        b.customPackageUrls,
				CorePackageCustomization: b.packageCustomization,
			},
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("creating localbuild resource: %w", err)
	}

	// Create GiteaProvider CR for v2 architecture
	setupLog.Info("Creating giteaprovider resource")
	if err := b.createGiteaProvider(ctx, kubeClient); err != nil {
		return fmt.Errorf("creating giteaprovider resource: %w", err)
	}

	// Create Platform CR that references GiteaProvider
	setupLog.Info("Creating platform resource")
	if err := b.createPlatform(ctx, kubeClient); err != nil {
		return fmt.Errorf("creating platform resource: %w", err)
	}

	select {
	case mgrErr := <-managerExit:
		if mgrErr != nil {
			return mgrErr
		}
	case <-ctx.Done():
		return nil
	}
	return nil
}

func isBuildCustomizationSpecEqual(s1, s2 v1alpha1.BuildCustomizationSpec) bool {
	// probably ok to use cmp.Equal but keeping it simple for now
	return s1.Protocol == s2.Protocol &&
		s1.Host == s2.Host &&
		s1.IngressHost == s2.IngressHost &&
		s1.Port == s2.Port &&
		s1.UsePathRouting == s2.UsePathRouting &&
		s1.SelfSignedCert == s2.SelfSignedCert &&
		s1.StaticPassword == s2.StaticPassword
}

// createGiteaProvider creates a GiteaProvider CR
func (b *Build) createGiteaProvider(ctx context.Context, kubeClient client.Client) error {
	// Ensure gitea namespace exists
	if err := k8s.EnsureNamespace(ctx, kubeClient, util.GiteaNamespace); err != nil {
		return fmt.Errorf("ensuring gitea namespace: %w", err)
	}

	giteaProvider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name + "-gitea",
			Namespace: util.GiteaNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, kubeClient, giteaProvider, func() error {
		giteaProvider.Spec = v1alpha2.GiteaProviderSpec{
			Namespace: util.GiteaNamespace,
			Version:   "1.24.3",
			AdminUser: v1alpha2.GiteaAdminUser{
				Username:     "giteaAdmin",
				Email:        "admin@" + b.cfg.Host,
				AutoGenerate: true,
			},
		}
		return nil
	})

	return err
}

// createPlatform creates a Platform CR that references the GiteaProvider
func (b *Build) createPlatform(ctx context.Context, kubeClient client.Client) error {
	platform := &v1alpha2.Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name + "-platform",
			Namespace: "default",
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, kubeClient, platform, func() error {
		platform.Spec = v1alpha2.PlatformSpec{
			Domain: b.cfg.Host,
			Components: v1alpha2.PlatformComponents{
				GitProviders: []v1alpha2.ProviderReference{
					{
						Name:      b.name + "-gitea",
						Kind:      "GiteaProvider",
						Namespace: util.GiteaNamespace,
					},
				},
			},
		}
		return nil
	})

	return err
}
