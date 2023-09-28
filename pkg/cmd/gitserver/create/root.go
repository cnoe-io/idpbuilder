package create

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"git.autodesk.com/forge-cd-services/idpbuilder/api/v1alpha1"
	"git.autodesk.com/forge-cd-services/idpbuilder/globals"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/apps"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/controllers"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/docker"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/k8s"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/kind"
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
	buildName  string
	serverName string
	sourcePath string

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
	CreateCmd.PersistentFlags().StringVar(&buildName, "buildName", "localdev", "Name for build (Prefix for kind cluster name, pod names, etc)")
	CreateCmd.PersistentFlags().StringVar(&serverName, "serverName", "", "Name of gitserver, must be unique within a build (name)")
	CreateCmd.PersistentFlags().StringVar(&sourcePath, "source", "", "Path to directory to use as gitserver source")

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

func create(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer ctxCancel()

	// Build a docker client
	dockerClient, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	// Build a kube client
	scheme := k8s.GetScheme()
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		setupLog.Error(err, "Building kubeconfig from kind cluster")
		return err
	}
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Creating kubernetes client")
		return err
	}

	// Create controller manager
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "Creating controller manager")
		return err
	}

	// Run controllers
	managerExit := make(chan error)
	if err := controllers.RunControllers(ctx, mgr, managerExit, ctxCancel); err != nil {
		setupLog.Error(err, "Running controllers")
		return err
	}

	// Build and push image
	srcFS := os.DirFS(sourcePath)
	imageTag := fmt.Sprintf("localhost:%d/%s-%s-%s", kind.ExposedRegistryPort, globals.ProjectName, buildName, serverName)
	imageID, err := apps.BuildAppsImage(ctx, dockerClient, []string{imageTag}, map[string]string{}, srcFS)
	if err != nil {
		return err
	}
	if imageID == nil {
		return fmt.Errorf("failed to get image id after build")
	}

	regImgId, err := apps.PushImage(ctx, dockerClient, imageTag)
	if err != nil {
		return err
	}
	if regImgId == nil {
		return fmt.Errorf("failed to get registry image id after push")
	}

	// Create GitServer k8s resource
	gitServer := v1alpha1.GitServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serverName,
			Namespace: globals.GetProjectNamespace(buildName),
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, kubeClient, &gitServer, func() error {
		gitServer.Spec = v1alpha1.GitServerSpec{
			Source: v1alpha1.GitServerSource{
				Image: fmt.Sprintf("%s@%s", imageTag, *regImgId),
			},
		}
		return nil
	}); err != nil {
		return err
	}

	err = <-managerExit
	close(managerExit)
	return err
}
