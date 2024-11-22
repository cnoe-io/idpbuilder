package get

import (
	"fmt"

	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/cluster"
)

var ClustersCmd = &cobra.Command{
	Use:          "clusters",
	Short:        "Get idp clusters",
	Long:         ``,
	RunE:         list,
	PreRunE:      preClustersE,
	SilenceUsage: true,
}

func preClustersE(cmd *cobra.Command, args []string) error {
	return helpers.SetLogger()
}

func list(cmd *cobra.Command, args []string) error {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())
	clusters, err := provider.List()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	for _, c := range clusters {
		fmt.Println(c)
	}
	return nil
}
