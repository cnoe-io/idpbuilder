package delete

import (
	"flag"

	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/kind/pkg/cluster"
)

var (
	// Flags
	buildName string
)

var DeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete an IDP cluster",
	Long:    ``,
	RunE:    delete,
	PreRunE: preDeleteE,
}

func init() {
	DeleteCmd.PersistentFlags().StringVar(&buildName, "build-name", "localdev", "Name of the kind cluster to be deleted.")

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(zapfs)
	DeleteCmd.Flags().AddGoFlagSet(zapfs)
}

func preDeleteE(cmd *cobra.Command, args []string) error {
	return helpers.SetLogger()
}

func delete(cmd *cobra.Command, args []string) error {
	detectOpt, err := cluster.DetectNodeProvider()
	if err != nil {
		return err
	}
	provider := cluster.NewProvider(detectOpt)
	if err := provider.Delete(buildName, ""); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %q", buildName)
	}
	return nil
}
