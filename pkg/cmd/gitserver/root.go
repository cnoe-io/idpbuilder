package gitserver

import (
	"fmt"
	"os"

	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/cmd/gitserver/create"
	"github.com/spf13/cobra"
)

var GitServerCmd = &cobra.Command{
	Use:   "gitserver",
	Short: "Interact with GitServers",
	Long:  "",
}

func init() {
	GitServerCmd.AddCommand(create.CreateCmd)
}

func Execute() {
	if err := GitServerCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
