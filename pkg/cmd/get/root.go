package get

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "get information from the cluster",
	Long:  ``,
	RunE:  exportE,
}

var (
	packages     []string
	outputFormat string
)

func init() {
	GetCmd.AddCommand(ClustersCmd)
	GetCmd.AddCommand(SecretsCmd)
	GetCmd.PersistentFlags().StringSliceVarP(&packages, "packages", "p", []string{}, "names of packages.")
	GetCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format. json or yaml.")
	GetCmd.PersistentFlags().StringVarP(&helpers.KubeConfigPath, "kubeconfig", "", "", "kube config file Path.")
}

func exportE(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("specify subcommand")
}
