package export

import (
	"github.com/spf13/cobra"
)

var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "export information from the cluster",
	Long:  ``,
	RunE:  exportE,
}

func init() {
	ExportCmd.AddCommand(SecretsCmd)
}

func exportE(cmd *cobra.Command, args []string) error {
	return nil
}
