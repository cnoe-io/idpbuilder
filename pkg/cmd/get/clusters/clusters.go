package clusters

import (
	"flag"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/kind/pkg/cluster"
)

var ClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Get idp clusters",
	Long:  ``,
	RunE:  list,
}

func init() {
	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(zapfs)
	ClustersCmd.Flags().AddGoFlagSet(zapfs)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func list(cmd *cobra.Command, args []string) error {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())
	clusters, err := provider.List()
	if err != nil {
		return errors.Wrapf(err, "failed to list clusters")
	}

	for _, cluster := range clusters {
		fmt.Println(cluster)
	}
	return nil
}
