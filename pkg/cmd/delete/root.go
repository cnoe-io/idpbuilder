package delete

import (
	"flag"
	"fmt"
	"os"

	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	// Flags
	buildName string
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an IDP cluster",
	Long:  ``,
	RunE:  delete,
}

func init() {
	DeleteCmd.PersistentFlags().StringVar(&buildName, "buildName", "localdev", "Name for build (Prefix for kind cluster name, pod names, etc).")

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(zapfs)
	DeleteCmd.Flags().AddGoFlagSet(zapfs)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func delete(cmd *cobra.Command, args []string) error {
	if buildName == "" {
		fmt.Print("Must specify buildName\n")
		os.Exit(1)
	}

	cluster, err := kind.NewCluster(buildName, "", "", "", "")
	if err != nil {
		return err
	}
	if err := cluster.Delete(); err != nil {
		return err
	}
	return nil
}
