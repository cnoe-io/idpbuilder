package get

import (
	"fmt"

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
	GetCmd.AddCommand(SecretsCmd)
	GetCmd.PersistentFlags().StringSliceVarP(&packages, "packages", "p", []string{}, "names of packages.")
	GetCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format. json or yaml.")
}

func exportE(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("specify subcommand")
}
