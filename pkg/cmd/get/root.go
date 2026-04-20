package get

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/util"
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
	GetCmd.AddCommand(PackagesCmd)
	GetCmd.AddCommand(CertificateCmd)
	GetCmd.PersistentFlags().StringSliceVarP(&packages, "packages", "p", []string{}, "names of packages.")
	GetCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table (default if not specified), json or yaml.")
	GetCmd.PersistentFlags().StringVarP(&util.KubeConfigPath, "kubeconfig", "", "", "kube config file Path.")
}

func exportE(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("specify subcommand")
}
