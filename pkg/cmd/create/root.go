package create

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/client-go/util/homedir"
)

var (
	// Flags
	recreateCluster   bool
	buildName         string
	kubeVersion       string
	extraPortsMapping string
	kindConfigPath    string
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "(Re)Create an IDP cluster",
	Long:  ``,
	RunE:  create,
}

func init() {
	CreateCmd.PersistentFlags().BoolVar(&recreateCluster, "recreate", false, "Delete cluster first if it already exists.")
	CreateCmd.PersistentFlags().StringVar(&buildName, "buildName", "localdev", "Name for build (Prefix for kind cluster name, pod names, etc).")
	CreateCmd.PersistentFlags().StringVar(&kubeVersion, "kubeVersion", "v1.26.3", "Version of the kind kubernetes cluster to create.")
	CreateCmd.PersistentFlags().StringVar(&extraPortsMapping, "extraPorts", "", "List of extra ports to expose on the docker container and kubernetes cluster as nodePort (e.g. \"22:32222,9090:39090,etc\").")
	CreateCmd.PersistentFlags().StringVar(&kindConfigPath, "kindConfig", "", "Path of the kind config file to be used instead of the default.")

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(zapfs)
	CreateCmd.Flags().AddGoFlagSet(zapfs)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func create(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer ctxCancel()

	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if buildName == "" {
		fmt.Print("Must specify buildName\n")
		os.Exit(1)
	}

	b := build.NewBuild(buildName, kubeVersion, kubeConfigPath, kindConfigPath, extraPortsMapping, k8s.GetScheme(), ctxCancel)

	if err := b.Run(ctx, recreateCluster); err != nil {
		return err
	}
	return nil
}
