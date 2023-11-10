package create

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/apps"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/gitserver"
	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	// Flags
	serverName     string
	sourcePath     string
	kubeConfigPath string

	setupLog = ctrl.Log.WithName("setup")
)

var CreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a GitServer",
	Long:    ``,
	PreRunE: preCreate,
	RunE:    create,
}

func init() {
	CreateCmd.PersistentFlags().StringVar(&serverName, "serverName", "", "Name of gitserver")
	CreateCmd.PersistentFlags().StringVar(&sourcePath, "source", "", "Path to directory to use as gitserver source")
	CreateCmd.PersistentFlags().StringVar(&kubeConfigPath, "kubeConfigPath", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Path to Kubernetes config.")

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(zapfs)
	CreateCmd.Flags().AddGoFlagSet(zapfs)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func preCreate(cmd *cobra.Command, args []string) error {
	if sourcePath == "" {
		return fmt.Errorf("source argument required")
	}
	if serverName == "" {
		return fmt.Errorf("serverName argument required")
	}
	return nil
}

func createPackageGitServer(ctx context.Context, kubeClient client.Client, name string, imageId string) (*v1alpha1.GitServer, error) {
	resource := v1alpha1.GitServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, kubeClient, &resource, func() error {
		resource.Spec = v1alpha1.GitServerSpec{
			Source: v1alpha1.GitServerSource{
				Image: imageId,
			},
		}
		return nil
	}); err != nil {
		setupLog.Error(err, "Creating package gitserver")
		return nil, err
	}

	key := client.ObjectKeyFromObject(&resource)
	for {
		setupLog.Info("Waiting for gitserver to become available")
		if err := kubeClient.Get(ctx, key, &resource); err != nil {
			setupLog.Error(err, "Getting gitserver object when waiting")
			return nil, err
		}

		if !resource.Status.DeploymentAvailable {
			time.Sleep(time.Second)
		} else {
			setupLog.Info("GitServer available")
			break
		}
	}

	return &resource, nil
}

func createPackageImage(ctx context.Context, name string) (*string, error) {
	dockerClient, err := docker.GetDockerClient()
	if err != nil {
		return nil, err
	}

	// Create gitserver image
	crossplaneFS := os.DirFS(sourcePath)
	imageTag := fmt.Sprintf("localhost:%d/%s-gitserver-%s", kind.ExposedRegistryPort, name)
	if _, err := apps.BuildAppsImage(ctx, dockerClient, []string{imageTag}, map[string]string{}, crossplaneFS); err != nil {
		setupLog.Error(err, "Building package image")
		return nil, err
	}

	// Push crossplane gitserver image
	regImgId, err := apps.PushImage(ctx, dockerClient, imageTag)
	if err != nil {
		setupLog.Error(err, "Pushing package image")
		return nil, err
	}
	if regImgId == nil {
		err = fmt.Errorf("nil img id when pushing crossplane image")
		setupLog.Error(err, "Pushing package image")
		return nil, err
	}
	imageId := fmt.Sprintf("%s@%s", imageTag, *regImgId)
	return &imageId, nil
}

func create(cmd *cobra.Command, args []string) error {
	ctx := ctrl.SetupSignalHandler()
	scheme := k8s.GetScheme()
	managerExit := make(chan error)
	ctx, ctxCancel := context.WithCancel(ctx)
	defer ctxCancel()

	// Get kube config
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		setupLog.Error(err, "Error building kubeconfig from kind cluster")
		return err
	}

	// Get kube client
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Error creating kubernetes client")
		return err
	}

	// Create controller manager
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "Creating package controller manager")
		return err
	}

	// Run GitServer controller
	appsFS, err := apps.GetAppsFS()
	if err != nil {
		setupLog.Error(err, "unable to find srv dir in apps fs")
		return err
	}
	if err := (&gitserver.GitServerReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Content: appsFS,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create gitserver controller")
		return err
	}

	// Start our manager in another goroutine
	setupLog.Info("starting manager")
	go func() {
		if err := mgr.Start(ctx); err != nil {
			setupLog.Error(err, "problem running manager")
			managerExit <- err
		}
		managerExit <- nil
	}()

	imgId, err := createPackageImage(ctx, serverName)
	if err != nil {
		return err
	}

	_, err = createPackageGitServer(ctx, kubeClient, serverName, *imgId)
	if err != nil {
		return err
	}

	err = <-managerExit
	close(managerExit)
	return err
}
