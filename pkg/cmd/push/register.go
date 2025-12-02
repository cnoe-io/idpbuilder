package push

import "github.com/spf13/cobra"

// AddToRoot adds the push command to the root command.
// Called from pkg/cmd/root.go during initialization.
func AddToRoot(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PushCmd)
}
