package delete

import (
	"fmt"

	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/cluster"
)

var (
	// Flags
	name string
)

var DeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Delete an IDP cluster",
	Long:         ``,
	RunE:         deleteE,
	PreRunE:      preDeleteE,
	SilenceUsage: true,
}

func init() {
	DeleteCmd.PersistentFlags().StringVar(&name, "name", "localdev", "Name of the kind cluster to be deleted.")
}

func preDeleteE(cmd *cobra.Command, args []string) error {
	return helpers.SetLogger()
}

func deleteE(cmd *cobra.Command, args []string) error {
	logger := helpers.CmdLogger
	logger.Info("deleting cluster", "clusterName", name)
	detectOpt, err := util.DetectKindNodeProvider()
	if err != nil {
		return err
	}

	provider := cluster.NewProvider(cluster.ProviderWithLogger(kind.KindLoggerFromLogr(&logger)), detectOpt)
	if err := provider.Delete(name, ""); err != nil {
		return fmt.Errorf("failed to delete cluster %s: %w", name, err)
	}
	return nil
}
