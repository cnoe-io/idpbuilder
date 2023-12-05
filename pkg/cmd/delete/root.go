package delete

import (
	"flag"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/kind/pkg/cluster"
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
	DeleteCmd.PersistentFlags().StringVar(&buildName, "build-name", "localdev", "Name of the kind cluster to be deleted.")

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(zapfs)
	DeleteCmd.Flags().AddGoFlagSet(zapfs)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func delete(cmd *cobra.Command, args []string) error {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())
	if err := provider.Delete(buildName, ""); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %q", buildName)
	}
	return nil
}
