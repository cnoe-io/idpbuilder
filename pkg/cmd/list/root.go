package list

import (
	"flag"
	"fmt"

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

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List idp clusters",
	Long:  ``,
	RunE:  delete,
}

func init() {
	ListCmd.PersistentFlags().StringVar(&buildName, "build-name", "localdev", "Name of the kind cluster to be deleted.")

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(zapfs)
	ListCmd.Flags().AddGoFlagSet(zapfs)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func delete(cmd *cobra.Command, args []string) error {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())
	clusters, err := provider.List()
	if err != nil {
		return errors.Wrapf(err, "failed to delete cluster %q", buildName)
	}

	fmt.Printf("Clusters: %v\n", clusters)
	return nil
}
